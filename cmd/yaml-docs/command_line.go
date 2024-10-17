package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/blakyaks/yaml-docs/pkg/document"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version string

func possibleLogLevels() []string {
	levels := make([]string, 0)

	for _, l := range log.AllLevels {
		levels = append(levels, l.String())
	}

	return levels
}

func initializeCli() {
	logLevelName := viper.GetString("log-level")
	logLevel, err := log.ParseLevel(logLevelName)
	if err != nil {
		log.Errorf("Failed to parse provided log level %s: %s", logLevelName, err)
		os.Exit(1)
	}

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(logLevel)
}

func newYamlDocsCommand(run func(cmd *cobra.Command, args []string)) (*cobra.Command, error) {
	command := &cobra.Command{
		Use:     "yaml-docs",
		Short:   "yaml-docs automatically generates markdown documentation from YAML configuration files",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			configSearchRoot, _ := cmd.Flags().GetString("config-search-root")
			configFiles, _ := cmd.Flags().GetStringSlice("config-file")

			if configSearchRoot != "" && len(configFiles) > 0 {
				log.Error("config-search-root and values-file are mutually exclusive.")
				os.Exit(1)
			}

			if configSearchRoot == "" && len(configFiles) == 0 {
				log.Error("One of config-search-root or values-file must be provided.")
				os.Exit(1)
			}

			run(cmd, args)
		},
	}

	logLevelUsage := fmt.Sprintf("Level of logs that should printed, one of (%s)", strings.Join(possibleLogLevels(), ", "))
	command.PersistentFlags().Bool("ignore-non-descriptions", false, "ignore values without a comment, these values will not be included in the README")
	command.PersistentFlags().Bool("multiple-output-files", false, "if set each config-file will render its own README template using the output-file-prefix format")
	command.PersistentFlags().Bool("skip-version-footer", false, "if set, the yaml-docs version footer will not be shown in the default README template")
	command.PersistentFlags().Bool("disable-section-inheritance", false, "if set, sections will not be inherited during document processing")
	command.PersistentFlags().BoolP("documentation-strict-mode", "x", false, "Fail the generation of docs if there are undocumented values")
	command.PersistentFlags().BoolP("dry-run", "d", false, "don't actually render any markdown files just print to stdout passed")
	command.PersistentFlags().StringP("config-search-root", "c", "", "directory to search recursively for configuration files, mutually exclusive with values-file")
	command.PersistentFlags().StringP("header-file", "H", ".document-header.md", "The external header content file that will be prepended to the standard template output. If the file does not exist templates will render without a custom header.")
	command.PersistentFlags().StringP("ignore-file", "i", ".yamldocsignore", "The filename to use as an ignore file to exclude configuration directories and files")
	command.PersistentFlags().StringP("log-level", "l", "info", logLevelUsage)
	command.PersistentFlags().StringP("output-file-prefix", "p", "README-%s.md", "The prefix format used when multiple output files are specified")
	command.PersistentFlags().StringP("output-file", "o", "README.md", "markdown file path where rendered documentation will be written")
	command.PersistentFlags().StringP("sort-values-order", "s", document.AlphaNumSortOrder, fmt.Sprintf("order in which to sort the values table (\"%s\" or \"%s\")", document.AlphaNumSortOrder, document.FileSortOrder))
	command.PersistentFlags().StringSliceP("config-file", "f", []string{}, "yaml configuration file to be parsed into values table. Can be specified multiple times, mutually exclusive with config-search-root")
	command.PersistentFlags().StringSliceP("documentation-strict-ignore-absent-regex", "z", []string{".*service\\.type", ".*image\\.repository", ".*image\\.tag"}, "A comma separate values which are allowed not to be documented in strict mode")
	command.PersistentFlags().StringSliceP("documentation-strict-ignore-absent", "y", []string{"service.type", "image.repository", "image.tag"}, "A comma separate values which are allowed not to be documented in strict mode")
	command.PersistentFlags().StringSliceP("template-files", "t", []string{"README.md.gotmpl"}, "gotemplate file paths relative to each configuration directory from which documentation will be generated")

	command.SetVersionTemplate(`{{printf "%s" .Version}}`)

	viper.AutomaticEnv()
	viper.SetEnvPrefix("YAML_DOCS")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	err := viper.BindPFlags(command.PersistentFlags())

	return command, err
}

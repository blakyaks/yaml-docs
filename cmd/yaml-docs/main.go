package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/blakyaks/yaml-docs/pkg/config"
	"github.com/blakyaks/yaml-docs/pkg/document"
)

func main() {
	command, err := newYamlDocsCommand(yamlDocs)
	if err != nil {
		log.Errorf("Failed to create the CLI commander: %s", err)
		os.Exit(1)
	}

	if err := command.Execute(); err != nil {
		log.Errorf("Failed to start the CLI: %s", err)
		os.Exit(1)
	}
}

// parallelProcessIterable runs the visitFn function on each element of the iterable, using
// parallelism number of worker goroutines. The iterable may be a slice or a map. In the case of a
// map, the argument passed to visitFn will be the key.
func parallelProcessIterable(iterable interface{}, parallelism int, visitFn func(elem interface{})) {
	workChan := make(chan interface{})

	wg := &sync.WaitGroup{}
	wg.Add(parallelism)

	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()
			for elem := range workChan {
				visitFn(elem)
			}
		}()
	}

	iterableValue := reflect.ValueOf(iterable)

	if iterableValue.Kind() == reflect.Map {
		for _, key := range iterableValue.MapKeys() {
			workChan <- key.Interface()
		}
	} else {
		sliceLen := iterableValue.Len()
		for i := 0; i < sliceLen; i++ {
			workChan <- iterableValue.Index(i).Interface()
		}
	}

	close(workChan)
	wg.Wait()
}

func getDocumentationParsingConfigFromArgs() (config.DocumentationParsingConfig, error) {
	var regexps []*regexp.Regexp
	regexpStrings := viper.GetStringSlice("documentation-strict-ignore-absent-regex")
	for _, item := range regexpStrings {
		regex, err := regexp.Compile(item)
		if err != nil {
			return config.DocumentationParsingConfig{}, err
		}
		regexps = append(regexps, regex)
	}
	return config.DocumentationParsingConfig{
		StrictMode:                 viper.GetBool("documentation-strict-mode"),
		AllowedMissingValuePaths:   viper.GetStringSlice("documentation-strict-ignore-absent"),
		AllowedMissingValueRegexps: regexps,
	}, nil
}

// ProcessConfigPaths processes config paths in parallel and returns a map of DocumentationInfo keyed by config path.
func processConfigPaths(configPaths []string, parallelism int) (map[string]config.DocumentationInfo, error) {

	// Common part
	templateFiles := viper.GetStringSlice("template-files")
	log.Debugf("Rendering from optional template files [%s]", strings.Join(templateFiles, ", "))

	documentationInfoByConfigPath := make(map[string]config.DocumentationInfo, len(configPaths))
	documentationInfoByConfigPathMu := &sync.Mutex{}
	documentationParsingConfig, err := getDocumentationParsingConfigFromArgs()
	if err != nil {
		return nil, fmt.Errorf("error parsing the linting config: %w", err)
	}

	// Process config paths
	parallelProcessIterable(configPaths, parallelism, func(elem interface{}) {
		configPath := elem.(string)

		if !path.IsAbs(configPath) {
			cwd, err := os.Getwd()
			if err != nil {
				log.Warnf("Error getting working directory: %v", err)
				return
			}
			configPath = filepath.Join(cwd, configPath)
		}

		info, err := config.ParseConfigPath(configPath, documentationParsingConfig)
		if err != nil {
			if parseError, ok := err.(*config.ConfigParseError); ok {
				log.Warnf("Configuration parse error at %s: %s", parseError.ConfigPath, parseError.Message)
				return
			} else {
				log.Warnf("Error parsing information for configuration directory %s, skipping: %s", configPath, err)
				return
			}
		}
		documentationInfoByConfigPathMu.Lock()
		documentationInfoByConfigPath[info.ConfigPath] = info
		documentationInfoByConfigPathMu.Unlock()
	})

	return documentationInfoByConfigPath, nil
}

func writeDocumentationMap(info map[string]config.DocumentationInfo, dryRun bool, parallelism int) {
	templateFiles := viper.GetStringSlice("template-files")
	skipVersionFooter := viper.GetBool("skip-version-footer")
	log.Debugf("Rendering from optional template files [%s]", strings.Join(templateFiles, ", "))

	parallelProcessIterable(info, parallelism, func(elem interface{}) {
		info := info[elem.(string)]
		document.PrintDocumentation(info, templateFiles, dryRun, version, skipVersionFooter)
	})
}

func writeDocumentation(info config.DocumentationInfo, dryRun bool) {
	templateFiles := viper.GetStringSlice("template-files")
	skipVersionFooter := viper.GetBool("skip-version-footer")
	log.Debugf("Rendering from optional template files [%s]", strings.Join(templateFiles, ", "))

	document.PrintDocumentation(info, templateFiles, dryRun, version, skipVersionFooter)
}

func yamlDocs(_ *cobra.Command, _ []string) {
	initializeCli()

	configSearchRoot := viper.GetString("config-search-root")
	configFiles := viper.GetStringSlice("config-file")
	createMultipleFiles := viper.GetBool("multiple-output-files")
	dryRun := viper.GetBool("dry-run")
	parallelism := runtime.NumCPU() * 2

	// On dry runs all output goes to stdout, and so as to not jumble things, generate serially.
	if dryRun {
		parallelism = 1
	}

	var info map[string]config.DocumentationInfo
	var err error

	if configSearchRoot != "" {
		info, err = processConfigPaths([]string{configSearchRoot}, parallelism)
	} else {
		info, err = processConfigPaths(configFiles, parallelism)
	}

	if err != nil {
		log.Fatal(err)
	}

	if len(info) == 0 {
		log.Warn("No YAML files were found, documentation will not be created.")
	} else {
		if createMultipleFiles {
			writeDocumentationMap(info, dryRun, parallelism)
		} else {
			combinedInfo := config.CombineDocumentationInfo(info)
			writeDocumentation(combinedInfo, dryRun)
		}
	}

}

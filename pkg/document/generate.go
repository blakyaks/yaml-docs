package document

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/blakyaks/yaml-docs/pkg/config"
	"github.com/blakyaks/yaml-docs/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getOutputFile(outputFile string, dryRun bool) (*os.File, error) {
	if dryRun {
		return os.Stdout, nil
	}

	f, err := os.Create(outputFile)

	if err != nil {
		return nil, err
	}

	return f, err
}

func PrintDocumentation(chartDocumentationInfo config.DocumentationInfo, templateFiles []string, dryRun bool, yamlDocsVersion string, skipVersionFooter bool) {
	log.Infof("Generating README Documentation for: %s", chartDocumentationInfo.ConfigPath)

	chartDocumentationTemplate, err := newChartDocumentationTemplate(templateFiles)

	if err != nil {
		log.Warnf("Error generating gotemplates: %s", err)
		return
	}

	chartTemplateDataObject, err := getChartTemplateData(chartDocumentationInfo, yamlDocsVersion, skipVersionFooter)
	if err != nil {
		log.Warnf("Error generating template data: %s", err)
		return
	}

	f := viper.GetString("output-file")
	if viper.GetBool("multiple-output-files") {
		baseFilename := util.GetBaseFilename(chartDocumentationInfo.ConfigPath)
		outputFilePrefix := viper.GetString("output-file-prefix")
		f = fmt.Sprintf(outputFilePrefix, baseFilename)
	}

	outputFile, err := getOutputFile(f, dryRun)
	if err != nil {
		log.Warnf("Could not open chart README file %s", err)
		return
	}

	if !dryRun {
		defer outputFile.Close()
	}

	var output bytes.Buffer
	err = chartDocumentationTemplate.Execute(&output, chartTemplateDataObject)
	if err != nil {
		log.Warnf("Error generating documentation for chart: %s", err)
	}

	output = applyMarkDownFormat(output)
	_, err = output.WriteTo(outputFile)
	if err != nil {
		log.Warnf("Error generating documentation file for chart: %s", err)
	}
}

func applyMarkDownFormat(output bytes.Buffer) bytes.Buffer {
	outputString := output.String()
	re := regexp.MustCompile(` \n`)
	outputString = re.ReplaceAllString(outputString, "\n")

	re = regexp.MustCompile(`\n{3,}`)
	outputString = re.ReplaceAllString(outputString, "\n\n")

	output.Reset()
	output.WriteString(outputString)
	return output
}

package config_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/blakyaks/yaml-docs/pkg/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigParsingTestSuite struct {
	suite.Suite
}

func (suite *ConfigParsingTestSuite) SetupTest() {
	viper.Set("ignore-file", ".ignore")
	viper.SetConfigType("yaml")
}

func TestConfigParsingTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigParsingTestSuite))
}

func (suite *ConfigParsingTestSuite) TestNotFullyDocumentedChartStrictModeOff() {
	configPath := filepath.Join("test-fixtures", "full-template")
	_, err := config.ParseConfigPath(configPath, config.DocumentationParsingConfig{
		StrictMode: false,
	})
	suite.NoError(err)
}

func (suite *ConfigParsingTestSuite) TestNotFullyDocumentedChartStrictModeOn() {
	configPath := filepath.Join("test-fixtures", "full-template")
	_, err := config.ParseConfigPath(configPath, config.DocumentationParsingConfig{
		StrictMode: true,
	})
	expectedError := `values without documentation: 
controller
controller.name
controller.image
controller.image.repository
controller.image.tag
controller.extraVolumes
controller.extraVolumes.[0].name
controller.extraVolumes.[0].configMap
controller.extraVolumes.[0].configMap.name
controller.publishService
controller.service
controller.service.annotations
controller.service.annotations.external-dns.alpha.kubernetes.io/hostname
controller.service.type`
	suite.EqualError(err, expectedError)
}

func (suite *ConfigParsingTestSuite) TestNotFullyDocumentedChartStrictModeOnIgnores() {
	chartPath := filepath.Join("test-fixtures", "full-template")
	_, err := config.ParseConfigPath(chartPath, config.DocumentationParsingConfig{
		StrictMode: true,
		AllowedMissingValuePaths: []string{
			"controller",
			"controller.image",
			"controller.name",
			"controller.image.repository",
			"controller.image.tag",
			"controller.extraVolumes",
			"controller.extraVolumes.[0].name",
			"controller.extraVolumes.[0].configMap",
			"controller.extraVolumes.[0].configMap.name",
			"controller.publishService",
			"controller.service",
			"controller.service.annotations",
			"controller.service.annotations.external-dns.alpha.kubernetes.io/hostname",
			"controller.service.type",
		},
	})
	suite.NoError(err)
}

func (suite *ConfigParsingTestSuite) TestNotFullyDocumentedChartStrictModeOnIgnoresRegexp() {
	chartPath := filepath.Join("test-fixtures", "full-template")
	_, err := config.ParseConfigPath(chartPath, config.DocumentationParsingConfig{
		StrictMode: true,
		AllowedMissingValueRegexps: []*regexp.Regexp{
			regexp.MustCompile("controller.*"),
		},
	})
	suite.NoError(err)
}

func (suite *ConfigParsingTestSuite) TestFullyDocumentedChartStrictModeOn() {
	configPath := filepath.Join("test-fixtures", "fully-documented")
	_, err := config.ParseConfigPath(configPath, config.DocumentationParsingConfig{
		StrictMode: true,
	})
	suite.NoError(err)
}

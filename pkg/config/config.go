package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blakyaks/yaml-docs/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var valuesDescriptionRegex = regexp.MustCompile(`^\s*#\s*(.*)\s+--\s*(.*)$`)
var rawDescriptionRegex = regexp.MustCompile(`^\s*#\s+@raw`)
var commentContinuationRegex = regexp.MustCompile(`^\s*#(\s?)(.*)$`)
var defaultValueRegex = regexp.MustCompile(`^\s*# @default -- (.*)$`)
var valueFlagsRegex = regexp.MustCompile(`^(\(([^)]+)\)\s+)?((?:@\w+\s+)*)?(.*)$`)
var valueNotationTypeRegex = regexp.MustCompile(`^\s*#\s+@notationType\s+--\s+(.*)$`)
var sectionRegex = regexp.MustCompile(`^\s*# @section -- (.*)$`)
var exampleDescriptionRegex = regexp.MustCompile(`^\s*# @exampleDescription -- (.*)$`)
var exampleRegex = regexp.MustCompile(`^\s*# @example\s+(.*?)\s*-- (.*)$`)

type ParseError struct {
	ConfigPath string
	Message    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s - %s", e.Message, e.ConfigPath)
}

type ValueDescription struct {
	Description        string
	Default            string
	Section            string
	ValueType          string
	NotationType       string
	ExampleName        string
	ExampleDescription string
	Example            string
	Hidden             bool
	Required           bool
	Deprecated         bool
}

type DocumentationInfo struct {
	ConfigPath         string
	Values             *yaml.Node
	ValuesDescriptions map[string]ValueDescription
}

type DocumentationParsingConfig struct {
	StrictMode                 bool
	AllowedMissingValuePaths   []string
	AllowedMissingValueRegexps []*regexp.Regexp
}

// Main routine to enumerate a config directory contents
func ParseConfigPath(configDirectory string, documentationParsingConfig DocumentationParsingConfig) (DocumentationInfo, error) {
	var chartDocInfo DocumentationInfo
	var files []string

	ignoreFilename := viper.GetString("ignore-file")
	ignoreContext := util.NewIgnoreContext(ignoreFilename)

	err := filepath.Walk(configDirectory, enumerateYamlFiles(&files, &ignoreContext))
	if err != nil {
		log.Printf("Error walking through directory: %v", err)
		return chartDocInfo, err
	}

	if len(files) == 0 {
		log.Debugf("No YAML files were found in the path: %s.", configDirectory)
		return chartDocInfo, &ParseError{
			ConfigPath: configDirectory,
			Message:    "No YAML files were found in the path",
		}
	}

	// Get values data from configuration files
	chartValues, err := parseValues(files)
	if err != nil {
		return chartDocInfo, err
	}

	// Enumerate comments and descriptions from files
	chartDescriptions, err := parseValueDescriptions(files, chartValues, documentationParsingConfig)
	if err != nil {
		return chartDocInfo, err
	}

	chartDocInfo.ConfigPath = configDirectory
	chartDocInfo.Values = chartValues
	chartDocInfo.ValuesDescriptions = chartDescriptions

	return chartDocInfo, nil
}

// Helper function that merges multiple documentation info objects into one
func CombineDocumentationInfo(maps map[string]DocumentationInfo) DocumentationInfo {

	if len(maps) == 1 {
		for _, value := range maps {
			return value
		}
	}

	combined := DocumentationInfo{
		Values:             &yaml.Node{},
		ValuesDescriptions: make(map[string]ValueDescription),
	}

	for _, docInfo := range maps {
		combined.ConfigPath += docInfo.ConfigPath

		if docInfo.Values != nil {
			combined.Values = util.MergeYAMLNodes(combined.Values, docInfo.Values)
		}

		for k, v := range docInfo.ValuesDescriptions {
			combined.ValuesDescriptions[k] = v
		}
	}

	return combined
}

// Internal functions below
func getYamlFileContents(filename string) ([]byte, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, err
	}

	yamlFileContents, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return []byte(strings.Replace(string(yamlFileContents), "\r\n", "\n", -1)), nil
}

func shouldIgnore(path string, info os.FileInfo, ignoreContext *util.IgnoreContext) bool {
	if ignoreContext.ShouldIgnore(filepath.Dir(path), info) {
		return true
	} else {
		return ignoreContext.ShouldIgnore(path, info)
	}
}

func isErrorInReadingNecessaryFile(filePath string, loadError error) bool {
	if loadError != nil {
		if os.IsNotExist(loadError) {
			log.Warnf("Required configuration file %s missing. Skipping documentation for file", filePath)
			return true
		} else {
			log.Warnf("Error occurred in reading configuration file %s. Skipping documentation for file", filePath)
			return true
		}
	}

	return false
}

func removeIgnored(rootNode *yaml.Node, parentKind yaml.Kind) {
	newContent := make([]*yaml.Node, 0, len(rootNode.Content))
	for i := 0; i < len(rootNode.Content); i++ {
		node := rootNode.Content[i]
		if !strings.Contains(node.HeadComment, "@ignore") {
			removeIgnored(node, node.Kind)
			newContent = append(newContent, node)
		} else if parentKind == yaml.MappingNode {
			// for parentKind each yaml key is represented by two nodes
			i++
		}
	}
	rootNode.Content = newContent
}

func enumerateYamlFiles(files *[]string, ignoreContext *util.IgnoreContext) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml") {
			if !shouldIgnore(path, info, ignoreContext) {
				*files = append(*files, path)
			}
		}
		return nil
	}
}

func parseConfigFile(configFile string) (yaml.Node, error) {
	yamlFileContents, err := getYamlFileContents(configFile)

	var values yaml.Node
	if isErrorInReadingNecessaryFile(configFile, err) {
		return values, err
	}

	err = yaml.Unmarshal(yamlFileContents, &values)
	removeIgnored(&values, values.Kind)
	return values, err
}

func checkDocumentation(rootNode *yaml.Node, comments map[string]ValueDescription, config DocumentationParsingConfig) error {
	if len(rootNode.Content) == 0 {
		return nil
	}
	valuesWithoutDocs := collectValuesWithoutDoc(rootNode.Content[0], comments, make([]string, 0))
	valuesWithoutDocsAfterIgnore := make([]string, 0)
	for _, valueWithoutDoc := range valuesWithoutDocs {
		ignored := false
		for _, ignorableValuePath := range config.AllowedMissingValuePaths {
			ignored = ignored || valueWithoutDoc == ignorableValuePath
		}
		for _, ignorableValueRegexp := range config.AllowedMissingValueRegexps {
			ignored = ignored || ignorableValueRegexp.MatchString(valueWithoutDoc)
		}
		if !ignored {
			valuesWithoutDocsAfterIgnore = append(valuesWithoutDocsAfterIgnore, valueWithoutDoc)
		}
	}
	if len(valuesWithoutDocsAfterIgnore) > 0 {
		return fmt.Errorf("values without documentation: \n%s", strings.Join(valuesWithoutDocsAfterIgnore, "\n"))
	}
	return nil
}

func collectValuesWithoutDoc(node *yaml.Node, comments map[string]ValueDescription, currentPath []string) []string {
	valuesWithoutDocs := make([]string, 0)
	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode, valueNode := node.Content[i], node.Content[i+1]
			currentPath = append(currentPath, keyNode.Value)
			pathString := strings.Join(currentPath, ".")
			if _, ok := comments[pathString]; !ok {
				valuesWithoutDocs = append(valuesWithoutDocs, pathString)
			}

			childValuesWithoutDoc := collectValuesWithoutDoc(valueNode, comments, currentPath)
			valuesWithoutDocs = append(valuesWithoutDocs, childValuesWithoutDoc...)

			currentPath = currentPath[:len(currentPath)-1]
		}
	case yaml.SequenceNode:
		for i := 0; i < len(node.Content); i++ {
			valueNode := node.Content[i]
			currentPath = append(currentPath, fmt.Sprintf("[%d]", i))
			childValuesWithoutDoc := collectValuesWithoutDoc(valueNode, comments, currentPath)
			valuesWithoutDocs = append(valuesWithoutDocs, childValuesWithoutDoc...)
			currentPath = currentPath[:len(currentPath)-1]
		}
	}
	return valuesWithoutDocs
}

func parseConfigFileComments(configFile string, values *yaml.Node, lintingConfig DocumentationParsingConfig) (map[string]ValueDescription, error) {
	valuesFile, err := os.Open(configFile)

	if isErrorInReadingNecessaryFile(configFile, err) {
		return map[string]ValueDescription{}, err
	}

	defer valuesFile.Close()

	keyToDescriptions := make(map[string]ValueDescription)
	scanner := bufio.NewScanner(valuesFile)
	foundValuesComment := false
	commentLines := make([]string, 0)
	currentLineIdx := -1

	for scanner.Scan() {
		currentLineIdx++
		currentLine := scanner.Text()

		// If we've not yet found a values comment with a key name, try and find one on each line
		if !foundValuesComment {
			match := valuesDescriptionRegex.FindStringSubmatch(currentLine)
			if len(match) < 3 || match[1] == "" {
				continue
			}
			foundValuesComment = true
			commentLines = append(commentLines, currentLine)
			continue
		}

		// If we've already found a values comment, on the next line try and parse a comment continuation, a custom default value, or a section comment.
		// If we find continuations we can add them to the list and continue to the next line until we find a section comment or default value.
		// If we find a default value, we can add it to the list and continue to the next line. In the case we don't find one, we continue looking for a section comment.
		// When we eventually find a section comment, we add it to the list and conclude matching for the current key. If we don't find one, matching is also concluded.
		//
		// NOTE: This isn't readily enforced yet, because we can match the section comment and custom default value more than once and in another order, although this is just overwriting it.
		// Values comment, possible continuation, default value once or none then section comment once or none should be the preferred order.
		defaultCommentMatch := defaultValueRegex.FindStringSubmatch(currentLine)
		sectionCommentMatch := sectionRegex.FindStringSubmatch(currentLine)
		exampleDescriptionCommentMatch := exampleRegex.FindStringSubmatch(currentLine)
		exampleCommentMatch := exampleRegex.FindStringSubmatch(currentLine)
		commentContinuationMatch := commentContinuationRegex.FindStringSubmatch(currentLine)

		if len(exampleDescriptionCommentMatch) > 1 || len(exampleCommentMatch) > 1 || len(defaultCommentMatch) > 1 || len(sectionCommentMatch) > 1 || len(commentContinuationMatch) > 1 {
			commentLines = append(commentLines, currentLine)
			continue
		}

		// If we haven't continued by this point, we didn't match any of the comment formats we want, so we need to add
		// the in progress value to the map, and reset to looking for a new key
		key, description := ParseComment(commentLines)
		if key != "" {
			keyToDescriptions[key] = description
		}

		commentLines = make([]string, 0)
		foundValuesComment = false
	}

	if lintingConfig.StrictMode {
		err := checkDocumentation(values, keyToDescriptions, lintingConfig)
		if err != nil {
			return nil, err
		}
	}
	return keyToDescriptions, nil
}

func joinConfigFiles(valuesNodes []yaml.Node) yaml.Node {

	totalLen := 0
	for _, node := range valuesNodes {
		totalLen += len(node.Content)
	}

	mergedValuesContent := make([]*yaml.Node, totalLen)
	idx := 0
	for _, node := range valuesNodes {
		idx += copy(mergedValuesContent[idx:], node.Content)
	}

	mergedValues := yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: mergedValuesContent,
	}

	return mergedValues
}

func parseValues(configFileNames []string) (*yaml.Node, error) {

	valuesNodes := make([]yaml.Node, len(configFileNames))

	for idx, valuesFile := range configFileNames {
		values, err := parseConfigFile(valuesFile)
		if err != nil {
			log.Warnf("Error parsing values from file: %s", valuesFile)
			values = yaml.Node{}
		}

		if err != nil {
			log.Warnf("Error parsing values from file: %s", valuesFile)
			values = yaml.Node{}
		}

		valuesNodes[idx] = values
	}

	mergedValues := joinConfigFiles(valuesNodes)
	return &mergedValues, nil
}

func parseValueDescriptions(configFileNames []string, values *yaml.Node, lintingConfig DocumentationParsingConfig) (map[string]ValueDescription, error) {

	valuesNodes := make([]map[string]ValueDescription, len(configFileNames))

	for idx, valuesFile := range configFileNames {
		values, err := parseConfigFileComments(valuesFile, values, lintingConfig)
		if err != nil {
			log.Warnf("Error parsing comments from file: %s", valuesFile)
			return nil, err
		}
		valuesNodes[idx] = values
	}

	mergedValues := mergeValueDescriptionMaps(valuesNodes...)

	return mergedValues, nil
}

func mergeValueDescriptionMaps(maps ...map[string]ValueDescription) map[string]ValueDescription {

	// Make a new map to hold the result
	result := make(map[string]ValueDescription)

	// Iterate over the list of maps
	for _, m := range maps {
		// Copy each map to the result
		for key, value := range m {
			result[key] = value
		}
	}

	return result
}

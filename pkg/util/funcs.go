package util

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

func FuncMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	f["toYaml"] = toYAML
	f["fromYaml"] = fromYAML
	f["include"] = includeFileContent
	f["trimLead"] = trimLeadingSpace
	f["toYamlCodeBlock"] = toYamlCodeBlock
	return f
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// fromYAML converts a YAML document into a map[string]interface{}.
//
// This is not a general-purpose YAML parser, and will not parse all valid
// YAML documents. Additionally, because its intended use is within templates
// it tolerates errors. It will insert the returned error message string into
// m["Error"] in the returned map.
func fromYAML(str string) map[string]interface{} {
	m := map[string]interface{}{}

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		m["Error"] = err.Error()
	}
	return m
}

// includeFileContent will read the filename provided and include it in the template.
// It is designed to import HTML or markdown content from outside of the core template
// and mirrors the behaviour of the include function in terraform-docs.
func includeFileContent(filename string) string {

	var fullPath string

	if filepath.IsAbs(filename) {
		fullPath = filename
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return ""
		}
		fullPath = filepath.Join(cwd, filename)
	}

	// Read the file, return empty string if not found
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return ""
	}

	return strings.TrimSuffix(string(content), "\n")
}

// Returns the string with any leading whitespace removed
// Can be used from templates using the format {{ .Property | trimLead }}
func trimLeadingSpace(str string) string {
	return strings.TrimLeftFunc(str, unicode.IsSpace)
}

// Returns the string wrapped in a YAML-format code block
// Can be used from templates using the format {{ .Property | toYamlCodeBlock }}
func toYamlCodeBlock(str string) string {
	return "```yaml\n" + str + "\n```"
}

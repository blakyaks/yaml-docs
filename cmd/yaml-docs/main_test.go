package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/blakyaks/yaml-docs/pkg/document"
)

var _ viper.FlagValueSet = &testFlagSet{}

type testFlagSet map[string]interface{}

func (s testFlagSet) VisitAll(fn func(viper.FlagValue)) {
	for k, v := range s {
		flagVal := &testFlagValue{
			name:  k,
			value: fmt.Sprintf("%v", v),
		}
		switch v.(type) {
		case bool:
			flagVal.typ = "bool"
		default:
			flagVal.typ = "string"
		}
		fn(flagVal)
	}
}

var _ viper.FlagValue = &testFlagValue{}

type testFlagValue struct {
	name  string
	value string
	typ   string
}

func (v *testFlagValue) HasChanged() bool {
	return false
}

func (v *testFlagValue) Name() string {
	return v.name
}

func (v *testFlagValue) ValueString() string {
	return v.value
}

func (v *testFlagValue) ValueType() string {
	return v.typ
}

func TestSkipsVersionFooter(t *testing.T) {
	// Ensure the skip-version-footer flag is true, meaning the README should not contain the footer.
	if err := viper.BindFlagValues(testFlagSet{
		"config-search-root":  filepath.Join("test_fixtures", "skip-version-footer"),
		"template-files":      "README.md.gotmpl",
		"output-file":         filepath.Join("test_fixtures", "skip-version-footer", "README.md"),
		"ignore-file":         ".yamldocsignore",
		"log-level":           "warn",
		"sort-values-order":   document.AlphaNumSortOrder,
		"skip-version-footer": true,
	}); err != nil {
		t.Fatal(err)
	}

	// Clean up the generated README after testing.
	readmePath := filepath.Join("test_fixtures", "skip-version-footer", "README.md")
	t.Cleanup(func() {
		if err := os.Remove(readmePath); err != nil {
			t.Fatal(err)
		}
	})

	// Generate the README, setting the helm-docs version number.
	version = "1.2.3"
	yamlDocs(nil, nil)

	// Confirm the helm-docs version is not present.
	docBytes, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(docBytes)
	if strings.Contains(doc, "Autogenerated from chart metadata using [yaml-docs v1.2.3]") {
		t.Errorf("generated documentation should not contain the yaml-docs version footer, got %s", doc)
	}
}

func TestIncludesVersionFooter(t *testing.T) {
	// Ensure the skip-version-footer flag is true, meaning the README must contain the footer.
	if err := viper.BindFlagValues(testFlagSet{
		"config-search-root":  filepath.Join("test_fixtures", "skip-version-footer"),
		"template-files":      "README.md.gotmpl",
		"output-file":         filepath.Join("test_fixtures", "skip-version-footer", "README.md"),
		"ignore-file":         ".yamldocsignore",
		"log-level":           "warn",
		"sort-values-order":   document.AlphaNumSortOrder,
		"skip-version-footer": false,
	}); err != nil {
		t.Fatal(err)
	}

	// Clean up the generated README after testing.
	readmePath := filepath.Join("test_fixtures", "skip-version-footer", "README.md")
	t.Cleanup(func() {
		if err := os.Remove(readmePath); err != nil {
			t.Fatal(err)
		}
	})

	// Generate the README, setting the helm-docs version number.
	version = "1.2.3"
	yamlDocs(nil, nil)

	// Confirm the helm-docs version is present
	docBytes, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(docBytes)
	if !strings.Contains(doc, "Autogenerated from configuration metadata using [yaml-docs v1.2.3]") {
		t.Errorf("generated documentation must contain the yaml-docs version footer, got %s", doc)
	}
}

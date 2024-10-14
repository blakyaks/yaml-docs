package document

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/blakyaks/yaml-docs/pkg/config"
)

type valueRow struct {
	Key                string
	Type               string
	NotationType       string
	AutoDefault        string
	Default            string
	AutoDescription    string
	Description        string
	Section            string
	Example            string
	ExampleDescription string
	Column             int
	LineNumber         int
}

type chartTemplateData struct {
	config.DocumentationInfo
	YamlDocsVersion   string
	Values            []valueRow
	Sections          sections
	Files             files
	SkipVersionFooter bool
	DocumentHeader    string
}

type sections struct {
	DefaultSection section
	Sections       []section
}

type section struct {
	SectionName  string
	SectionItems []valueRow
	Examples     []example
}

type example struct {
	ExampleName string
	Description string
	CodeBlock   string
}

func sortValueRowsByOrder(valueRows []valueRow, sortOrder string) {
	sort.Slice(valueRows, func(i, j int) bool {
		// Sort the remaining values within the same section using the configured sort order.
		switch sortOrder {
		case FileSortOrder:
			if valueRows[i].LineNumber == valueRows[j].LineNumber {
				return valueRows[i].Column < valueRows[j].Column
			}
			return valueRows[i].LineNumber < valueRows[j].LineNumber
		case AlphaNumSortOrder:
			return valueRows[i].Key < valueRows[j].Key
		default:
			panic("cannot get here")
		}
	})
}

func sortValueRows(valueRows []valueRow) {
	sortOrder := viper.GetString("sort-values-order")

	if sortOrder != FileSortOrder && sortOrder != AlphaNumSortOrder {
		log.Warnf("Invalid sort order provided %s, defaulting to %s", sortOrder, AlphaNumSortOrder)
		sortOrder = AlphaNumSortOrder
	}

	sortValueRowsByOrder(valueRows, sortOrder)
}

func sortSectionedValueRows(sectionedValueRows sections) {
	sortOrder := viper.GetString("sort-values-order")

	if sortOrder != FileSortOrder && sortOrder != AlphaNumSortOrder {
		log.Warnf("Invalid sort order provided %s, defaulting to %s", sortOrder, AlphaNumSortOrder)
		sortOrder = AlphaNumSortOrder
	}

	sortValueRowsByOrder(sectionedValueRows.DefaultSection.SectionItems, sortOrder)

	for _, section := range sectionedValueRows.Sections {
		sortValueRowsByOrder(section.SectionItems, sortOrder)
	}
}

func getUnsortedValueRows(document *yaml.Node, descriptions map[string]config.ValueDescription) ([]valueRow, error) {

	// Handle empty values file case.
	if document.Kind == 0 {
		return nil, nil
	}

	if document.Kind != yaml.DocumentNode {
		return nil, fmt.Errorf("invalid node kind supplied: %d", document.Kind)
	}

	if document.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("values file must resolve to a map (was %d)", document.Content[0].Kind)
	}

	var allValueRows []valueRow

	for _, contentNode := range document.Content {
		valueRows, err := createValueRowsFromField("", nil, contentNode, descriptions, true)
		if err != nil {
			return nil, err
		}
		allValueRows = append(allValueRows, valueRows...)
	}

	return allValueRows, nil
}

func getSectionedValueRows(valueRows []valueRow) sections {
	var valueRowsSectionSorted sections
	valueRowsSectionSorted.DefaultSection = section{
		SectionName:  "Other Values",
		SectionItems: []valueRow{},
		Examples:     []example{},
	}

	for _, row := range valueRows {
		if row.Section == "" {
			valueRowsSectionSorted.DefaultSection.SectionItems = append(valueRowsSectionSorted.DefaultSection.SectionItems, row)

			if row.Example != "" {
				valueRowsSectionSorted.DefaultSection.Examples = append(valueRowsSectionSorted.DefaultSection.Examples, example{
					ExampleName: row.Key,
					Description: row.ExampleDescription,
					CodeBlock:   row.Example,
				})
			}

			continue
		}

		containsSection := false
		for i, section := range valueRowsSectionSorted.Sections {
			if section.SectionName == row.Section {
				containsSection = true
				valueRowsSectionSorted.Sections[i].SectionItems = append(valueRowsSectionSorted.Sections[i].SectionItems, row)
				break
			}
			if row.Example != "" {
				valueRowsSectionSorted.Sections[i].Examples = append(valueRowsSectionSorted.Sections[i].Examples, example{
					ExampleName: row.Key,
					Description: row.ExampleDescription,
					CodeBlock:   row.Example,
				})
			}
		}

		if !containsSection {
			var examples []example
			// Only create the examples slice if there is a valid example
			if row.Example != "" {
				examples = []example{{
					ExampleName: row.Key,
					Description: row.ExampleDescription,
					CodeBlock:   row.Example,
				}}
			}
			// Append a new section regardless of whether examples exist
			valueRowsSectionSorted.Sections = append(valueRowsSectionSorted.Sections, section{
				SectionName:  row.Section,
				SectionItems: []valueRow{row},
				Examples:     examples,
			})
		}
	}

	return valueRowsSectionSorted
}

func getChartTemplateData(info config.DocumentationInfo, yamlDocsVersion string, skipVersionFooter bool) (chartTemplateData, error) {
	valuesTableRows, err := getUnsortedValueRows(info.Values, info.ValuesDescriptions)
	if err != nil {
		return chartTemplateData{}, err
	}

	if viper.GetBool("ignore-non-descriptions") {
		valuesTableRows = removeRowsWithoutDescription(valuesTableRows)
	}

	sortValueRows(valuesTableRows)
	valueRowsSectionSorted := getSectionedValueRows(valuesTableRows)
	sortSectionedValueRows(valueRowsSectionSorted)

	documentHeaderFile := viper.GetString("header-file")

	// TODO: Add examples

	return chartTemplateData{
		DocumentationInfo: info,
		YamlDocsVersion:   yamlDocsVersion,
		Values:            valuesTableRows,
		Sections:          valueRowsSectionSorted,
		SkipVersionFooter: skipVersionFooter,
		DocumentHeader:    documentHeaderFile,
	}, nil
}

func removeRowsWithoutDescription(valuesTableRows []valueRow) []valueRow {

	var valuesTableRowsWithoutDescription []valueRow
	for i := range valuesTableRows {
		if valuesTableRows[i].AutoDescription != "" || valuesTableRows[i].Description != "" {
			valuesTableRowsWithoutDescription = append(valuesTableRowsWithoutDescription, valuesTableRows[i])
		}
	}
	return valuesTableRowsWithoutDescription
}

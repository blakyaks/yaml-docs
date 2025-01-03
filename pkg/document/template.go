package document

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blakyaks/yaml-docs/pkg/util"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

const defaultDocumentationTemplate = `{{ include .DocumentHeader }}
{{- if .CreateToc }}
{{ template "config.sectionToc" . }}
{{- end }}
{{ template "config.examplesSection" . }}
{{ template "config.valuesSection" . }}
{{- if not .SkipVersionFooter }}
{{ template "yaml-docs.versionFooter" . }}
{{- end }}
`

func getDocumentationTemplate(templateFiles []string) (string, error) {
	templateFilesForChart := make([]string, 0)

	var templateNotFound bool

	for _, templateFile := range templateFiles {
		fullTemplatePath := templateFile

		if _, err := os.Stat(fullTemplatePath); os.IsNotExist(err) {
			log.Debugf("Did not find template file %s, using default template", templateFile)

			templateNotFound = true
			continue
		}

		templateFilesForChart = append(templateFilesForChart, fullTemplatePath)
	}

	log.Debugf("Using template files %s", templateFiles)
	allTemplateContents := make([]byte, 0)
	for _, templateFileForChart := range templateFilesForChart {
		templateContents, err := os.ReadFile(templateFileForChart)
		if err != nil {
			return "", err
		}
		allTemplateContents = append(allTemplateContents, templateContents...)
	}

	if templateNotFound {
		allTemplateContents = append(allTemplateContents, []byte(defaultDocumentationTemplate)...)
	}

	return string(allTemplateContents), nil
}

func getDocumentationTemplates(templateFiles []string) ([]string, error) {
	documentationTemplate, err := getDocumentationTemplate(templateFiles)

	if err != nil {
		log.Errorf("Failed to read documentation templates %s: %s", templateFiles, err)
		return nil, err
	}

	return []string{
		getSectionToc(),
		getValuesTableTemplates(),
		getYamlDocsVersionTemplates(),
		getGlobalExamplesTemplates(),
		documentationTemplate,
	}, nil
}

func newChartDocumentationTemplate(templateFiles []string) (*template.Template, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	documentationTemplate := template.New(filepath.Base(cwd))
	documentationTemplate.Funcs(util.FuncMap())
	goTemplateList, err := getDocumentationTemplates(templateFiles)

	if err != nil {
		return nil, err
	}

	for _, t := range goTemplateList {
		_, err := documentationTemplate.Parse(t)

		if err != nil {
			return nil, err
		}
	}

	return documentationTemplate, nil
}

func getGlobalExamplesTemplates() string {
	s := strings.Builder{}
	s.WriteString(`{{ define "config.examplesHeader" }}## Examples{{ end }}`)
	s.WriteString(`{{ define "config.examplesSection" }}`)
	s.WriteString("{{ if .Sections.DefaultSection.Examples }}")
	s.WriteString("{{ if .Sections.DefaultSection.SectionItems }}")
	s.WriteString(`{{ template "config.examplesHeader" . }}`)
	s.WriteString("{{- end }}")
	s.WriteString("\n\n")
	s.WriteString("{{ range .Sections.DefaultSection.Examples }}")
	s.WriteString("\n### {{ .ExampleName }}\n\n")
	s.WriteString("{{ if .Description }}{{ .Description }}\n\n{{ end }}")
	s.WriteString("{{ .CodeBlock | trimLead | toYamlCodeBlock }}\n\n")
	s.WriteString("{{ end }}")
	s.WriteString("-----------------\n")
	s.WriteString("{{ end }}")
	s.WriteString("{{ end }}")

	return s.String()
}

func getYamlDocsVersionTemplates() string {
	s := strings.Builder{}
	s.WriteString(`{{ define "yaml-docs.version" }}{{ if .YamlDocsVersion }}{{ .YamlDocsVersion }}{{ end }}{{ end }}`)
	s.WriteString(`{{ define "yaml-docs.versionFooter" }}`)
	s.WriteString("{{ if .YamlDocsVersion }}\n")
	s.WriteString("-----------------\n")
	s.WriteString("Autogenerated from configuration metadata using [yaml-docs v{{ .YamlDocsVersion }}](https://github.com/blakyaks/yaml-docs/releases/v{{ .YamlDocsVersion }})")
	s.WriteString("{{ end }}")
	s.WriteString("{{ end }}")

	return s.String()
}

func getSectionToc() string {
	s := strings.Builder{}
	s.WriteString(`{{ define "config.sectionToc" }}`)
	if viper.GetBool("no-section-page-breaks") {
		s.WriteString("\n-----------------\n\n")
	} else {
		s.WriteString("\n")
		s.WriteString("<div style=\"page-break-after: always;\"></div>")
		s.WriteString("\n\n")
	}
	s.WriteString("\n## Contents\n\n")
	s.WriteString("{{ range .Sections.Sections }}")
	s.WriteString("- {{ .SectionName | toMarkdownLink }}")
	s.WriteString("{{ end }}")
	s.WriteString("{{ if .Sections.DefaultSection.SectionItems }}")
	s.WriteString("- {{ .Sections.DefaultSection.SectionName | toMarkdownLink }}")
	s.WriteString("{{- end }}")
	if viper.GetBool("no-section-page-breaks") {
		s.WriteString("\n-----------------\n\n")
	} else {
		s.WriteString("\n")
		s.WriteString("<div style=\"page-break-after: always;\"></div>")
		s.WriteString("\n\n")
	}
	s.WriteString(`{{ end }}`)
	return s.String()
}

func getValuesTableTemplates() string {
	s := strings.Builder{}
	s.WriteString(`{{ define "config.valuesHeader" }}## Values{{ end }}`)
	s.WriteString(`{{ define "config.valuesTable" }}`)
	s.WriteString("{{ if .Sections.Sections }}")
	s.WriteString("{{ range .Sections.Sections }}")
	s.WriteString("\n")
	s.WriteString("\n### {{ .SectionName }}\n")
	s.WriteString("\n")
	s.WriteString("{{ if .Description }}{{ .Description }}\n\n{{ end }}")
	s.WriteString("{{ if .Examples }}")
	s.WriteString("#### Examples\n\n")
	s.WriteString("{{ range .Examples }}")
	s.WriteString("\n##### {{ .ExampleName }}\n\n")
	s.WriteString("{{ if .Description }}{{ .Description }}\n\n{{ end }}")
	s.WriteString("{{ .CodeBlock | trimLead | toYamlCodeBlock }}\n\n")
	s.WriteString("{{ end }}")
	s.WriteString("{{- end }}")
	s.WriteString("{{ if .Examples }}#### Values\n\n{{ end }}")
	s.WriteString("| Key | Type | Required | Default | Description |\n")
	s.WriteString("|-----|------|----------|---------|-------------|\n")
	s.WriteString("  {{- range .SectionItems }}")
	s.WriteString("  {{- if not .Hidden }}")
	s.WriteString("\n| {{ if .Experimental }}<span style='cursor: help;' title='Experimental'>✨</span>{{ end }}{{ if .Deprecated }}<span style='cursor: help;' title='Deprecated'>⚠️</span>{{ end }} {{ .Key }} | {{ .Type }} | {{ if .Required }}**{{ .Required }}**{{ else }}{{ .Required }}{{ end }} | {{ if .Default }}{{ .Default }}{{ else }}{{ .AutoDefault }}{{ end }} | {{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }} |")
	s.WriteString("  {{- end }}")
	s.WriteString("  {{- end }}")
	s.WriteString("{{ if .SectionBreak }}\n\n<div style=\"page-break-after: always;\"></div>{{ end }}")
	s.WriteString("{{- end }}")

	// Default section
	s.WriteString("{{ if .Sections.DefaultSection.SectionItems }}")
	s.WriteString("\n\n")
	s.WriteString("\n### {{ .Sections.DefaultSection.SectionName }}\n")
	s.WriteString("\n")
	s.WriteString("| Key | Type | Required | Default | Description |\n")
	s.WriteString("|-----|------|----------|---------|-------------|\n")
	s.WriteString("  {{- range .Sections.DefaultSection.SectionItems }}")
	s.WriteString("  {{- if not .Hidden }}")
	s.WriteString("\n| {{ if .Experimental }}<span style='cursor: help;' title='Experimental'>✨</span>{{ end }}{{ if .Deprecated }}<span style='cursor: help;' title='Deprecated'>⚠️</span>{{ end }} {{ .Key }} | {{ .Type }} | {{ if .Required }}**{{ .Required }}**{{ else }}{{ .Required }}{{ end }} | {{ if .Default }}{{ .Default }}{{ else }}{{ .AutoDefault }}{{ end }} | {{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }} |")
	s.WriteString("  {{- end }}")
	s.WriteString("  {{- end }}")
	s.WriteString("{{ end }}")
	s.WriteString("{{ else }}")
	s.WriteString("| Key | Type | Required | Default | Description |\n")
	s.WriteString("|-----|------|----------|---------|-------------|\n")
	s.WriteString("  {{- range .Values }}")
	s.WriteString("\n| {{ if .Experimental }}<span style='cursor: help;' title='Experimental'>✨</span>{{ end }}{{ if .Deprecated }}<span style='cursor: help;' title='Deprecated'>⚠️</span>{{ end }} {{ .Key }} | {{ .Type }} | {{ if .Required }}**{{ .Required }}**{{ else }}{{ .Required }}{{ end }} | {{ if .Default }}{{ .Default }}{{ else }}{{ .AutoDefault }}{{ end }} | {{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }} |")
	s.WriteString("  {{- end }}")
	s.WriteString("{{ end }}")
	s.WriteString("{{ end }}")
	s.WriteString(`{{ template "config.examplesHeader" . }}`)
	s.WriteString(`{{ template "config.examplesSection" . }}`)
	s.WriteString(`{{ define "config.valuesSection" }}`)
	s.WriteString("{{ if and .Sections.DefaultSection.SectionItems .Sections.DefaultSection.Examples }}")
	s.WriteString(`{{ template "config.valuesHeader" . }}`)
	s.WriteString("{{ end }}")
	s.WriteString("{{ if .Values }}")
	s.WriteString("\n\n")
	s.WriteString(`{{ template "config.valuesTable" . }}`)
	s.WriteString("{{ end }}")
	s.WriteString("{{- end }}")

	// For HTML tables
	s.WriteString(`
{{ define "config.valueDefaultColumnRender" }}
{{- $defaultValue := (default .Default .AutoDefault)  -}}
{{- $notationType := .NotationType }}
{{- if (and (hasPrefix "` + "`" + `" $defaultValue) (hasSuffix "` + "`" + `" $defaultValue) ) -}}
{{- $defaultValue = (toPrettyJson (fromJson (trimAll "` + "`" + `" (default .Default .AutoDefault) ) ) ) -}}
{{- $notationType = "json" }}
{{- end -}}
<pre lang="{{ $notationType }}">
{{- if (eq $notationType "tpl" ) }}
{{ .Key }}: |
{{- $defaultValue | nindent 2 }}
{{- else }}
{{ $defaultValue }}
{{- end }}
</pre>
{{ end }}

{{ define "config.valuesTableHtml" }}
{{ if .Sections.Sections }}
{{- range .Sections.Sections }}
<h3>{{- .SectionName }}</h3>
<table>
	<thead>
		<th>Key</th>
		<th>Type</th>
		<th>Required</th>
		<th>Default</th>
		<th>Description</th>
	</thead>
	<tbody>
	{{- range .SectionItems }}
		<tr>
			<td>{{ .Key }}</td>
			<td>{{ .Type }}</td>
			<td>{{ .Required }}</td>
			<td>{{ template "config.valueDefaultColumnRender" . }}</td>
			<td>{{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }}</td>
		</tr>
	{{- end }}
	</tbody>
</table>
{{- end }}
{{ if .Sections.DefaultSection.SectionItems }}
<h3>{{- .Sections.DefaultSection.SectionName }}</h3>
<table>
	<thead>
		<th>Key</th>
		<th>Type</th>
		<th>Required</th>
		<th>Default</th>
		<th>Description</th>
	</thead>
	<tbody>
	{{- range .Sections.DefaultSection.SectionItems }}
	<tr>
		<td>{{ .Key }}</td>
		<td>{{ .Type }}</td>
		<td>{{ .Required }}</td>
		<td>{{ template "config.valueDefaultColumnRender" . }}</td>
		<td>{{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }}</td>
	</tr>
	{{- end }}
	</tbody>
</table>
{{ end }}
{{ else }}
<table>
	<thead>
		<th>Key</th>
		<th>Type</th>
		<th>Required</th>
		<th>Default</th>
		<th>Description</th>
	</thead>
	<tbody>
	{{- range .Values }}
		<tr>
			<td>{{ .Key }}</td>
			<td>{{ .Type }}</td>
			<td>{{ .Required }}</td>
			<td>{{ template "config.valueDefaultColumnRender" . }}</td>
			<td>{{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }}</td>
		</tr>
	{{- end }}
	</tbody>
</table>
{{ end }}
{{ end }}

{{ define "config.valuesSectionHtml" }}
{{ if .Sections }}
{{ template "config.valuesHeader" . }}
{{ template "config.valuesTableHtml" . }}
{{ end }}
{{ end }}
		`)

	return s.String()
}

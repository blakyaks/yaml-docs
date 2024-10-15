package document

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blakyaks/yaml-docs/pkg/util"

	log "github.com/sirupsen/logrus"
)

const defaultDocumentationTemplate = `{{ include .DocumentHeader }}

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
	exampleSectionBuilder := strings.Builder{}
	exampleSectionBuilder.WriteString(`{{ define "config.examplesHeader" }}## Examples{{ end }}`)
	exampleSectionBuilder.WriteString(`{{ define "config.examplesSection" }}`)
	exampleSectionBuilder.WriteString("{{ if .Sections.DefaultSection.Examples }}")
	exampleSectionBuilder.WriteString(`{{ template "config.examplesHeader" . }}`)
	exampleSectionBuilder.WriteString("\n\n")
	exampleSectionBuilder.WriteString("{{ range .Sections.DefaultSection.Examples }}")
	exampleSectionBuilder.WriteString("\n### {{ .ExampleName }}\n\n")
	exampleSectionBuilder.WriteString("{{ if .Description }}{{ .Description }}\n\n{{ end }}")
	exampleSectionBuilder.WriteString("{{ .CodeBlock | trimLead | toYamlCodeBlock }}\n\n")
	exampleSectionBuilder.WriteString("{{ end }}")
	exampleSectionBuilder.WriteString("{{ end }}")
	exampleSectionBuilder.WriteString("-----------------\n")
	exampleSectionBuilder.WriteString("{{ end }}")

	return exampleSectionBuilder.String()
}

func getYamlDocsVersionTemplates() string {
	versionSectionBuilder := strings.Builder{}
	versionSectionBuilder.WriteString(`{{ define "yaml-docs.version" }}{{ if .YamlDocsVersion }}{{ .YamlDocsVersion }}{{ end }}{{ end }}`)
	versionSectionBuilder.WriteString(`{{ define "yaml-docs.versionFooter" }}`)
	versionSectionBuilder.WriteString("{{ if .YamlDocsVersion }}\n")
	versionSectionBuilder.WriteString("-----------------\n")
	versionSectionBuilder.WriteString("Autogenerated from configuration metadata using [yaml-docs v{{ .YamlDocsVersion }}](https://github.com/blakyaks/yaml-docs/releases/v{{ .YamlDocsVersion }})")
	versionSectionBuilder.WriteString("{{ end }}")
	versionSectionBuilder.WriteString("{{ end }}")

	return versionSectionBuilder.String()
}

func getValuesTableTemplates() string {
	valuesSectionBuilder := strings.Builder{}
	valuesSectionBuilder.WriteString(`{{ define "config.valuesHeader" }}## Values{{ end }}`)

	valuesSectionBuilder.WriteString(`{{ define "config.valuesTable" }}`)
	valuesSectionBuilder.WriteString("{{ if .Sections.Sections }}")
	valuesSectionBuilder.WriteString("{{ range .Sections.Sections }}")
	valuesSectionBuilder.WriteString("\n")
	valuesSectionBuilder.WriteString("\n### {{ .SectionName }}\n")
	valuesSectionBuilder.WriteString("\n")
	valuesSectionBuilder.WriteString("{{ if .Examples }}")
	valuesSectionBuilder.WriteString("#### Examples\n\n")
	valuesSectionBuilder.WriteString("{{ range .Examples }}")
	valuesSectionBuilder.WriteString("\n#### {{ .ExampleName }}\n\n")
	valuesSectionBuilder.WriteString("{{ if .Description }}{{ .Description }}\n\n{{ end }}")
	valuesSectionBuilder.WriteString("{{ .CodeBlock | trimLead | toYamlCodeBlock }}\n\n")
	valuesSectionBuilder.WriteString("{{ end }}")
	valuesSectionBuilder.WriteString("{{- end }}")
	valuesSectionBuilder.WriteString("{{ if .Examples }}#### Values\n\n{{ end }}")
	valuesSectionBuilder.WriteString("| Key | Type | Required | Default | Description |\n")
	valuesSectionBuilder.WriteString("|-----|------|----------|---------|-------------|\n")
	valuesSectionBuilder.WriteString("  {{- range .SectionItems }}")
	valuesSectionBuilder.WriteString("  {{- if not .Hidden }}")
	valuesSectionBuilder.WriteString("\n| {{ .Key }} | {{ .Type }} | {{ if .Required }}**{{ .Required }}**{{ else }}{{ .Required }}{{ end }} | {{ if .Default }}{{ .Default }}{{ else }}{{ .AutoDefault }}{{ end }} | {{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }} |")
	valuesSectionBuilder.WriteString("  {{- end }}")
	valuesSectionBuilder.WriteString("  {{- end }}")
	valuesSectionBuilder.WriteString("{{- end }}")
	valuesSectionBuilder.WriteString("{{ if .Sections.DefaultSection.SectionItems }}")
	valuesSectionBuilder.WriteString("\n\n")
	valuesSectionBuilder.WriteString("-----------------\n")
	valuesSectionBuilder.WriteString("\n### {{ .Sections.DefaultSection.SectionName }}\n")
	valuesSectionBuilder.WriteString("\n")
	valuesSectionBuilder.WriteString("| Key | Type | Required | Default | Description |\n")
	valuesSectionBuilder.WriteString("|-----|------|----------|---------|-------------|\n")
	valuesSectionBuilder.WriteString("  {{- range .Sections.DefaultSection.SectionItems }}")
	valuesSectionBuilder.WriteString("  {{- if not .Hidden }}")
	valuesSectionBuilder.WriteString("\n| {{ .Key }} | {{ .Type }} | {{ if .Required }}**{{ .Required }}**{{ else }}{{ .Required }}{{ end }} | {{ if .Default }}{{ .Default }}{{ else }}{{ .AutoDefault }}{{ end }} | {{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }} |")
	valuesSectionBuilder.WriteString("  {{- end }}")
	valuesSectionBuilder.WriteString("  {{- end }}")
	valuesSectionBuilder.WriteString("{{ end }}")
	valuesSectionBuilder.WriteString("{{ else }}")
	valuesSectionBuilder.WriteString("| Key | Type | Default | Description |\n")
	valuesSectionBuilder.WriteString("|-----|------|---------|-------------|\n")
	valuesSectionBuilder.WriteString("  {{- range .Values }}")
	valuesSectionBuilder.WriteString("\n| {{ .Key }} | {{ .Type }} | {{ if .Default }}{{ .Default }}{{ else }}{{ .AutoDefault }}{{ end }} | {{ if .Description }}{{ .Description }}{{ else }}{{ .AutoDescription }}{{ end }} |")
	valuesSectionBuilder.WriteString("  {{- end }}")
	valuesSectionBuilder.WriteString("{{ end }}")
	valuesSectionBuilder.WriteString("{{ end }}")

	valuesSectionBuilder.WriteString(`{{ template "config.examplesHeader" . }}`)
	valuesSectionBuilder.WriteString(`{{ template "config.examplesSection" . }}`)
	valuesSectionBuilder.WriteString(`{{ define "config.valuesSection" }}`)
	valuesSectionBuilder.WriteString("{{ if .Values }}")
	valuesSectionBuilder.WriteString(`{{ template "config.valuesHeader" . }}`)
	valuesSectionBuilder.WriteString("\n\n")
	valuesSectionBuilder.WriteString(`{{ template "config.valuesTable" . }}`)
	valuesSectionBuilder.WriteString("{{ end }}")
	valuesSectionBuilder.WriteString("{{ end }}")

	// For HTML tables
	valuesSectionBuilder.WriteString(`
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

	return valuesSectionBuilder.String()
}

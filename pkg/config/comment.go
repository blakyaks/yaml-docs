package config

import (
	"strings"
)

const (
	PrefixComment = "# --"
)

func ParseComment(commentLines []string) (string, ValueDescription) {
	var valueKey string
	var c ValueDescription
	var docStartIdx int

	// Work around https://github.com/norwoodj/helm-docs/issues/96 by considering only
	// the last "group" of comment lines starting with '# --'.
	lastIndex := 0
	for i, v := range commentLines {
		if strings.HasPrefix(v, PrefixComment) {
			lastIndex = i
		}
	}
	if lastIndex > 0 {
		// If there's a non-zero last index, consider that alone.
		return ParseComment(commentLines[lastIndex:])
	}

	for i := range commentLines {
		match := valuesDescriptionRegex.FindStringSubmatch(commentLines[i])
		if len(match) < 3 {
			continue
		}

		valueKey = match[1]
		c.Description = match[2]
		docStartIdx = i
		break
	}

	flagTypeMatch := valueFlagsRegex.FindStringSubmatch(c.Description)
	if len(flagTypeMatch) > 0 {
		if flagTypeMatch[2] != "" {
			c.ValueType = flagTypeMatch[2]
		}
		c.Hidden = strings.Contains(flagTypeMatch[3], "@hidden")
		c.Required = strings.Contains(flagTypeMatch[3], "@required")
		c.Deprecated = strings.Contains(flagTypeMatch[3], "@deprecated")
		c.Experimental = strings.Contains(flagTypeMatch[3], "@experimental")
		c.Description = flagTypeMatch[4]
	}

	var isRaw = false
	var isExample = false
	var isExampleDescription = false
	var isSectionDescription = false
	var isRawExampleDescription = false
	var isRawSectionDescription = false

	for _, line := range commentLines[docStartIdx+1:] {
		rawFlagMatch := rawDescriptionRegex.FindStringSubmatch(line)
		defaultCommentMatch := defaultValueRegex.FindStringSubmatch(line)
		notationTypeCommentMatch := valueNotationTypeRegex.FindStringSubmatch(line)
		sectionDescriptionCommentMatch := sectionDescriptionRegex.FindStringSubmatch(line)
		sectionCommentMatch := sectionRegex.FindStringSubmatch(line)
		exampleDescriptionCommentMatch := exampleDescriptionRegex.FindStringSubmatch(line)
		exampleCommentMatch := exampleRegex.FindStringSubmatch(line)

		if !isRaw && len(rawFlagMatch) == 1 {
			isRaw = true
			continue
		}

		if len(defaultCommentMatch) > 1 {
			c.Default = defaultCommentMatch[1]
			continue
		}

		if len(notationTypeCommentMatch) > 1 {
			c.NotationType = notationTypeCommentMatch[1]
			continue
		}

		if len(sectionDescriptionCommentMatch) > 1 {
			if sectionDescriptionCommentMatch[1] == "@raw" {
				isRawSectionDescription = true
			}
			c.SectionDescription = sectionDescriptionCommentMatch[2]
			isSectionDescription = true
			continue
		}

		if len(sectionCommentMatch) > 1 {
			c.Section = sectionCommentMatch[1]
			continue
		}

		if len(exampleDescriptionCommentMatch) > 1 {
			if exampleDescriptionCommentMatch[1] == "@raw" {
				isRawExampleDescription = true
			}
			c.ExampleDescription = exampleDescriptionCommentMatch[2]
			isExampleDescription = true
			continue
		}

		if len(exampleCommentMatch) > 1 {
			c.Example = exampleCommentMatch[2]
			c.ExampleName = exampleCommentMatch[1]
			isExample = true
			continue
		}

		commentContinuationMatch := commentContinuationRegex.FindStringSubmatch(line)

		if isExample && len(commentContinuationMatch) > 1 {
			c.Example += "\n" + commentContinuationMatch[2]
			continue
		}

		if isExampleDescription && len(commentContinuationMatch) > 1 {
			if isRawExampleDescription {
				c.ExampleDescription += "\n" + commentContinuationMatch[2]
			} else {
				c.ExampleDescription += " " + commentContinuationMatch[2]
			}
			continue
		}

		if isSectionDescription && len(commentContinuationMatch) > 1 {
			if isRawSectionDescription {
				c.SectionDescription += "\n" + commentContinuationMatch[2]
			} else {
				c.SectionDescription += " " + commentContinuationMatch[2]
			}
			continue
		}

		if isRaw {
			if len(commentContinuationMatch) > 1 {
				c.Description += "\n" + commentContinuationMatch[2]
			}
			continue
		} else {
			if len(commentContinuationMatch) > 1 {
				c.Description += " " + commentContinuationMatch[2]
			}
			isExample = false // Reset flags when not processing a continuation type
			isExampleDescription = false
			isSectionDescription = false
			isRawSectionDescription = false
			isRawExampleDescription = false
			continue
		}
	}

	return valueKey, c
}

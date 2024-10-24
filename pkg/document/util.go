package document

import (
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	AlphaNumSortOrder = "alphanum"
	FileSortOrder     = "file"
)

// The json library can only marshal maps with string keys, and so all of our lists and maps that go into documentation
// must be converted to have only string keys before marshalling
func convertConfigValuesToJsonable(values *yaml.Node) interface{} {
	switch values.Kind {
	case yaml.MappingNode:
		convertedMap := make(map[string]interface{})

		for i := 0; i < len(values.Content); i += 2 {
			k := values.Content[i]
			v := values.Content[i+1]
			convertedMap[k.Value] = convertConfigValuesToJsonable(v)
		}

		return convertedMap
	case yaml.SequenceNode:
		convertedList := make([]interface{}, 0)

		for _, v := range values.Content {
			convertedList = append(convertedList, convertConfigValuesToJsonable(v))
		}

		return convertedList
	case yaml.AliasNode:
		return convertConfigValuesToJsonable(values.Alias)
	case yaml.ScalarNode:
		switch values.Tag {
		case nullTag:
			return nil
		case strTag:
			fallthrough
		case timestampTag:
			return values.Value
		case intTag:
			var decodedValue int
			err := values.Decode(&decodedValue)
			if err != nil {
				log.Errorf("Failed to decode value from yaml node value %s", values.Value)
				return 0
			}
			return decodedValue
		case floatTag:
			var decodedValue float64
			err := values.Decode(&decodedValue)
			if err != nil {
				log.Errorf("Failed to decode value from yaml node value %s", values.Value)
				return 0
			}
			return decodedValue

		case boolTag:
			var decodedValue bool
			err := values.Decode(&decodedValue)
			if err != nil {
				log.Errorf("Failed to decode value from yaml node value %s", values.Value)
				return 0
			}
			return decodedValue
		}
	}

	return nil
}

// naturalLess compares two strings in a natural order.
func naturalLess(a, b string) bool {
	// Regular expression to extract numbers
	re := regexp.MustCompile(`(\d+)`)
	aIndexes := re.FindAllStringIndex(a, -1)
	bIndexes := re.FindAllStringIndex(b, -1)

	// Iterate over the slices of indexes and compare numerical values
	for i, aIndex := range aIndexes {
		if i >= len(bIndexes) {
			return false // b is shorter in number parts, a comes after b
		}
		bIndex := bIndexes[i]
		// Convert the substrings that are numbers to integers
		numA, _ := strconv.Atoi(a[aIndex[0]:aIndex[1]])
		numB, _ := strconv.Atoi(b[bIndex[0]:bIndex[1]])
		if numA != numB {
			return numA < numB // Compare as integers
		}
	}

	// If all numbers compared are equal, fall back to lexicographical order
	return a < b
}

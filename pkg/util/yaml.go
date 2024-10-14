package util

import "gopkg.in/yaml.v3"

func MergeYAMLNodes(nodes ...*yaml.Node) *yaml.Node {
	merged := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: make([]*yaml.Node, 0),
	}

	for _, node := range nodes {
		merged = mergeYAMLNodes(merged, node)
	}

	return merged
}

func mergeYAMLNodes(node1, node2 *yaml.Node) *yaml.Node {
	if node1.Kind == yaml.DocumentNode {
		node1 = node1.Content[0]
	}
	if node2.Kind == yaml.DocumentNode {
		node2 = node2.Content[0]
	}

	if node1.Kind != yaml.MappingNode || node2.Kind != yaml.MappingNode {
		// If either node is not a map, return node2
		return node2
	}

	// Make a map from node1 for easier access
	node1Map := make(map[string]*yaml.Node)
	for i := 0; i < len(node1.Content); i += 2 {
		key := node1.Content[i]
		value := node1.Content[i+1]
		node1Map[key.Value] = value
	}

	// Copy node1 to merged
	merged := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: append([]*yaml.Node(nil), node1.Content...),
	}

	// Iterate through node2 and add to merged
	for i := 0; i < len(node2.Content); i += 2 {
		key := node2.Content[i]
		value := node2.Content[i+1]

		// If this key is in node1, merge the values
		if node1Value, ok := node1Map[key.Value]; ok {
			// Find the index of the node1Value in merged.Content and replace it
			for j := 0; j < len(merged.Content); j += 2 {
				if merged.Content[j].Value == key.Value {
					merged.Content[j+1] = mergeYAMLNodes(node1Value, value)
					break
				}
			}
		} else {
			// If the key was not in node1, add it to merged
			merged.Content = append(merged.Content, key, value)
		}
	}

	return &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{merged},
	}
}

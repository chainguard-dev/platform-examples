package yamlhelpers

import "gopkg.in/yaml.v3"

// AddNode adds a node at the specified path
func AddNode(path []string, node *yaml.Node, add *yaml.Node) {
	if add == nil {
		return
	}

	current := node

	for i, key := range path {
		// If this is the last element of the path then add the node or
		// replace the existing node
		if i == len(path)-1 {
			// Check if key already exists and replace its value
			for j := 0; j < len(current.Content); j += 2 {
				if current.Content[j].Value == key {
					// Replace the existing value
					current.Content[j+1] = add
					return
				}
			}
			// Key doesn't exist, add new key-value pair
			current.Content = append(current.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: key},
				add,
			)
			return
		}

		// Otherwise, create the path (if it doesn't already exist)
		var next *yaml.Node
		for j := 0; j < len(current.Content); j += 2 {
			if current.Content[j].Value != key {
				continue
			}
			next = current.Content[j+1]
			break
		}
		if next == nil {
			// Create new intermediate mapping
			next = &yaml.Node{
				Kind:    yaml.MappingNode,
				Content: []*yaml.Node{},
			}
			current.Content = append(current.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: key},
				next,
			)
		}

		current = next
	}
}

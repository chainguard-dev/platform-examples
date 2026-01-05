package yamlhelpers

import "gopkg.in/yaml.v3"

// WalkNodeFn is called for each node by WalkNode
type WalkNodeFn func(path []string, node *yaml.Node) error

// WalkNode walks recursively through a yaml.Node, calling fn for each node.
func WalkNode(node *yaml.Node, fn WalkNodeFn) error {
	return walkNode([]string{}, node, fn)
}

func walkNode(path []string, node *yaml.Node, fn WalkNodeFn) error {
	if err := fn(path, node); err != nil {
		return err
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			value := node.Content[i+1]

			if err := walkNode(append(path, key.Value), value, fn); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if err := walkNode(path, child, fn); err != nil {
				return err
			}
		}
	}

	return nil
}

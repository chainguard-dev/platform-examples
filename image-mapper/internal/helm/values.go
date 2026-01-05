package helm

import (
	"context"
	"fmt"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/yamlhelpers"
	"gopkg.in/yaml.v3"
)

// MapValues extracts the image related values from a values file and maps them
// to Chainguard.
func MapValues(ctx context.Context, input []byte, opts ...mapper.Option) ([]byte, error) {
	m, err := NewMapper(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("constructing the new mapper: %w", err)
	}

	return mapValues(m, input)
}

// mapValues extracts the image related values from a values file and maps them
// to Chainguard with the provided mapper
func mapValues(m mapper.Mapper, input []byte) ([]byte, error) {
	var inputDoc yaml.Node
	if err := yaml.Unmarshal(input, &inputDoc); err != nil {
		return nil, fmt.Errorf("unmarshalling yaml: %w", err)
	}
	if len(inputDoc.Content) == 0 {
		return nil, fmt.Errorf("provided input document is empty")
	}
	inputNode := inputDoc.Content[0]

	// We'll write modified nodes to this node
	outputNode := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	// Walk the document recursively, adding image related fields to the
	// output node and mapping them to Chainguard images
	if err := yamlhelpers.WalkNode(inputNode, mapNode(m, []string{}, outputNode)); err != nil {
		return nil, fmt.Errorf("walking nodes: %w", err)
	}

	// Marshal the modified nodes to a new document
	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{outputNode},
	}
	output, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshalling output document: %w", err)
	}

	return output, nil
}

// mapNode returns a function that extracts image related fields from the input
// node and adds them to the output node, mapping the images to Chainguard where
// possible.
//
// It handles blocks like:
//
//	image:
//	  repository: ghcr.io/foo/bar
//	  tag: v0.0.1
//
//	OR
//
//	image:
//	  registry: ghcr.io
//	  repository: foo/bar
//	  tag: v0.0.1
//
//	OR
//
//	image:
//	  name: ghcr.io/foo/bar
//	  tag: v0.0.1
//
//	OR
//
//	image: ghcr.io/foo/bar:v0.0.1
func mapNode(m mapper.Mapper, yamlPath []string, output *yaml.Node) yamlhelpers.WalkNodeFn {
	return func(path []string, value *yaml.Node) error {
		if value.Kind != yaml.MappingNode {
			return nil
		}

		// Extract all the keys from the map that are typically
		// associated with an image
		var (
			image      *yaml.Node
			name       *yaml.Node
			repository *yaml.Node
			registry   *yaml.Node
			tag        *yaml.Node
		)
		for i := 0; i < len(value.Content); i += 2 {
			key := value.Content[i].Value
			value := value.Content[i+1]

			switch key {
			case "image":
				image = &yaml.Node{
					Kind:  value.Kind,
					Tag:   value.Tag,
					Value: value.Value,
				}
			case "name":
				name = &yaml.Node{
					Kind:  value.Kind,
					Tag:   value.Tag,
					Value: value.Value,
				}
			case "repository":
				repository = &yaml.Node{
					Kind:  value.Kind,
					Tag:   value.Tag,
					Value: value.Value,
				}
			case "registry":
				registry = &yaml.Node{
					Kind:  value.Kind,
					Tag:   value.Tag,
					Value: value.Value,
				}
			case "tag":
				tag = &yaml.Node{
					Kind:  value.Kind,
					Tag:   value.Tag,
					Value: value.Value,
				}
			}
		}

		// If we don't have one of repository, name or image then we
		// have no chance of figuring out the image mapping and we'll
		// skip over it.
		if !(hasValue(repository) || hasValue(name) || hasValue(image)) {
			return nil
		}

		// The key 'name' is too generic for us to assume it refers to
		// an image, so ignore maps with keys called 'name' unless
		// there are other signals that this is an image reference.
		//
		// For instance, if the map key is 'image', or we have a
		// registry/tag alongside the name.
		if hasValue(name) && !(path[len(path)-1] == "image" || registry != nil || tag != nil) {
			return nil
		}

		// Construct the image reference based on the fields
		// available
		img := ""
		if hasValue(name) {
			img = name.Value
		}
		if hasValue(image) {
			img = image.Value
		}
		if hasValue(repository) {
			img = repository.Value
		}
		if hasValue(registry) {
			img = fmt.Sprintf("%s/%s", registry.Value, img)
		}
		if hasValue(tag) {
			img = fmt.Sprintf("%s:%s", img, tag.Value)
		}

		// Map the constructed image reference to the equivalent
		// Chainguard image
		mapping, err := mapper.MapImage(m, img)
		if err == nil {
			// Modify the values to follow the mapped image. This
			// will ignore nodes that are nil.
			setValue(repository, mapping.Context().String())
			setValue(image, mapping.Context().String())
			setValue(name, mapping.Context().String())
			setValue(registry, mapping.Context().RegistryStr())

			// If there's no tag, then chances are image is a fully
			// qualified image reference
			if tag == nil {
				setValue(image, mapping.String())
			}

			// If the registry key exists, then the
			// repository shouldn't include the registry.
			if registry != nil {
				setValue(repository, mapping.Context().RepositoryStr())
				setValue(image, mapping.Context().RepositoryStr())
				setValue(name, mapping.Context().RepositoryStr())
			}

			// If the mapped tag is different to the tag in
			// the original values, then replace it.
			//
			// Otherwise, leave it alone so that the output values
			// don't include a specific tag have a better shot of
			// being compatible across chart version upgrades.
			if hasValue(tag) && tag.Value != mapping.Identifier() {
				setValue(tag, mapping.Identifier())
			}
		}

		// Create a new node and add all the modified values to it
		node := &yaml.Node{
			Kind:    yaml.MappingNode,
			Content: []*yaml.Node{},
		}
		if err != nil {
			node.HeadComment = fmt.Sprintf("Failed to map: %s: %s", img, err)
		}
		yamlhelpers.AddNode([]string{"registry"}, node, registry)
		yamlhelpers.AddNode([]string{"image"}, node, image)
		yamlhelpers.AddNode([]string{"name"}, node, name)
		yamlhelpers.AddNode([]string{"repository"}, node, repository)

		// Only include the tag if we modified it
		if tag != nil && tag.LineComment != "" {
			yamlhelpers.AddNode([]string{"tag"}, node, tag)
		}

		// Add the new node to the output values at the same path as the
		// input
		yamlhelpers.AddNode(append(yamlPath, path...), output, node)

		return nil
	}
}

// setValue sets the value of a scalar node
func setValue(node *yaml.Node, value string) {
	if node == nil {
		return
	}
	if node.Kind != yaml.ScalarNode {
		return
	}

	if node.LineComment == "" && node.Value != value {
		node.LineComment = fmt.Sprintf("Original: %s", node.Value)
	}

	node.Value = value
	node.Tag = "!!str"
}

// hasValue tells us whether a node has a value that we can try including in our
// mapping
func hasValue(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if node.Value == "" {
		return false
	}
	if node.Tag == "!!null" {
		return false
	}

	return true
}

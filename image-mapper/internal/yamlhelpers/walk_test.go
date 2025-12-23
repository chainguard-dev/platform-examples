package yamlhelpers

import (
	"errors"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWalkNode(t *testing.T) {
	testCases := []struct {
		name          string
		yaml          string
		expectedPaths []string
		expectedKinds []yaml.Kind
	}{
		{
			name: "simple mapping",
			yaml: `
key1: value1
key2: value2
`,
			expectedPaths: []string{
				"",
				"key1",
				"key2",
			},
			expectedKinds: []yaml.Kind{
				yaml.MappingNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
			},
		},
		{
			name: "nested mapping",
			yaml: `
parent:
  child1: value1
  child2: value2
`,
			expectedPaths: []string{
				"",
				"parent",
				"parent.child1",
				"parent.child2",
			},
			expectedKinds: []yaml.Kind{
				yaml.MappingNode,
				yaml.MappingNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
			},
		},
		{
			name: "sequence",
			yaml: `
items:
  - item1
  - item2
  - item3
`,
			expectedPaths: []string{
				"",
				"items",
				"items",
				"items",
				"items",
			},
			expectedKinds: []yaml.Kind{
				yaml.MappingNode,
				yaml.SequenceNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
			},
		},
		{
			name: "mixed nested structure",
			yaml: `
database:
  host: localhost
  port: 5432
  credentials:
    username: admin
    password: secret
servers:
  - name: server1
    ip: 192.168.1.1
  - name: server2
    ip: 192.168.1.2
`,
			expectedPaths: []string{
				"",
				"database",
				"database.host",
				"database.port",
				"database.credentials",
				"database.credentials.username",
				"database.credentials.password",
				"servers",
				"servers",
				"servers.name",
				"servers.ip",
				"servers",
				"servers.name",
				"servers.ip",
			},
			expectedKinds: []yaml.Kind{
				yaml.MappingNode,
				yaml.MappingNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
				yaml.MappingNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
				yaml.SequenceNode,
				yaml.MappingNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
				yaml.MappingNode,
				yaml.ScalarNode,
				yaml.ScalarNode,
			},
		},
		{
			name: "scalar only",
			yaml: `simple value`,
			expectedPaths: []string{
				"",
			},
			expectedKinds: []yaml.Kind{
				yaml.ScalarNode,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tc.yaml), &node)
			if err != nil {
				t.Fatalf("failed to unmarshal yaml: %v", err)
			}

			var paths []string
			var kinds []yaml.Kind

			walkFn := func(path []string, n *yaml.Node) error {
				pathStr := ""
				if len(path) > 0 {
					pathStr = path[0]
					for i := 1; i < len(path); i++ {
						pathStr += "." + path[i]
					}
				}
				paths = append(paths, pathStr)
				kinds = append(kinds, n.Kind)
				return nil
			}

			// Walk the content of the document node
			if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
				err = WalkNode(node.Content[0], walkFn)
			} else {
				err = WalkNode(&node, walkFn)
			}
			if err != nil {
				t.Fatalf("WalkNode returned error: %v", err)
			}

			if len(paths) != len(tc.expectedPaths) {
				t.Errorf("expected %d paths, got %d", len(tc.expectedPaths), len(paths))
			}

			for i := range paths {
				if i >= len(tc.expectedPaths) {
					break
				}
				if paths[i] != tc.expectedPaths[i] {
					t.Errorf("path[%d]: expected %q, got %q", i, tc.expectedPaths[i], paths[i])
				}
			}

			if len(kinds) != len(tc.expectedKinds) {
				t.Errorf("expected %d kinds, got %d", len(tc.expectedKinds), len(kinds))
			}

			for i := range kinds {
				if i >= len(tc.expectedKinds) {
					break
				}
				if kinds[i] != tc.expectedKinds[i] {
					t.Errorf("kind[%d]: expected %v, got %v", i, tc.expectedKinds[i], kinds[i])
				}
			}
		})
	}
}

func TestWalkNodeError(t *testing.T) {
	yamlContent := `
key1: value1
key2: value2
key3: value3
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	expectedErr := errors.New("test error")
	callCount := 0

	walkFn := func(path []string, n *yaml.Node) error {
		callCount++
		if callCount == 3 {
			return expectedErr
		}
		return nil
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		err = WalkNode(node.Content[0], walkFn)
	} else {
		err = WalkNode(&node, walkFn)
	}
	if err == nil {
		t.Error("expected error to be returned")
	}

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}

	if callCount != 3 {
		t.Errorf("expected walkFn to be called 3 times, was called %d times", callCount)
	}
}

func TestWalkNodeModifyValues(t *testing.T) {
	yamlContent := `
key1: value1
key2: value2
nested:
  key3: value3
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	walkFn := func(path []string, n *yaml.Node) error {
		if n.Kind == yaml.ScalarNode && n.Value != "" && len(path) > 0 {
			n.Value = "modified"
		}
		return nil
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		err = WalkNode(node.Content[0], walkFn)
	} else {
		err = WalkNode(&node, walkFn)
	}
	if err != nil {
		t.Fatalf("WalkNode returned error: %v", err)
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("failed to marshal modified yaml: %v", err)
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(out, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result["key1"] != "modified" {
		t.Errorf("expected key1 to be 'modified', got %v", result["key1"])
	}
	if result["key2"] != "modified" {
		t.Errorf("expected key2 to be 'modified', got %v", result["key2"])
	}

	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested to be a map")
	}
	if nested["key3"] != "modified" {
		t.Errorf("expected nested.key3 to be 'modified', got %v", nested["key3"])
	}
}

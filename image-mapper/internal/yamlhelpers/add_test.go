package yamlhelpers

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAddNode(t *testing.T) {
	testCases := []struct {
		name     string
		initial  string
		path     []string
		addValue string
		expected string
	}{
		{
			name:    "add to empty mapping",
			initial: `{}`,
			path:    []string{"newkey"},
			addValue: "newvalue",
			expected: `newkey: newvalue
`,
		},
		{
			name: "add new key to existing mapping",
			initial: `existing: value
`,
			path:    []string{"newkey"},
			addValue: "newvalue",
			expected: `existing: value
newkey: newvalue
`,
		},
		{
			name: "replace existing key",
			initial: `key: oldvalue
`,
			path:    []string{"key"},
			addValue: "newvalue",
			expected: `key: newvalue
`,
		},
		{
			name:    "create nested path",
			initial: `{}`,
			path:    []string{"level1", "level2", "level3"},
			addValue: "deepvalue",
			expected: `level1:
    level2:
        level3: deepvalue
`,
		},
		{
			name: "add to existing nested path",
			initial: `level1:
    existing: value
`,
			path:    []string{"level1", "newkey"},
			addValue: "newvalue",
			expected: `level1:
    existing: value
    newkey: newvalue
`,
		},
		{
			name: "replace in nested path",
			initial: `level1:
    level2:
        key: oldvalue
`,
			path:    []string{"level1", "level2", "key"},
			addValue: "newvalue",
			expected: `level1:
    level2:
        key: newvalue
`,
		},
		{
			name: "add sibling to nested structure",
			initial: `parent:
    child1: value1
`,
			path:    []string{"parent", "child2"},
			addValue: "value2",
			expected: `parent:
    child1: value1
    child2: value2
`,
		},
		{
			name: "create intermediate paths",
			initial: `existing: value
`,
			path:    []string{"new", "nested", "key"},
			addValue: "value",
			expected: `existing: value
new:
    nested:
        key: value
`,
		},
		{
			name: "add to root with multiple existing keys",
			initial: `key1: value1
key2: value2
key3: value3
`,
			path:    []string{"key4"},
			addValue: "value4",
			expected: `key1: value1
key2: value2
key3: value3
key4: value4
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tc.initial), &node)
			if err != nil {
				t.Fatalf("failed to unmarshal initial yaml: %v", err)
			}

			addNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: tc.addValue,
			}

			// The root node is a DocumentNode, we need to work with its first child (the MappingNode)
			if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
				AddNode(tc.path, node.Content[0], addNode)
			} else {
				t.Fatal("unexpected node structure")
			}

			out, err := yaml.Marshal(&node)
			if err != nil {
				t.Fatalf("failed to marshal result: %v", err)
			}

			// Unmarshal both to compare semantically rather than as strings
			var gotMap, expectedMap interface{}
			if err := yaml.Unmarshal(out, &gotMap); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if err := yaml.Unmarshal([]byte(tc.expected), &expectedMap); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			// Compare the semantic content
			gotYAML, _ := yaml.Marshal(gotMap)
			expectedYAML, _ := yaml.Marshal(expectedMap)
			if string(gotYAML) != string(expectedYAML) {
				t.Errorf("expected:\n%s\ngot:\n%s", string(expectedYAML), string(gotYAML))
			}
		})
	}
}

func TestAddNodeNil(t *testing.T) {
	initial := `key: value
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(initial), &node)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		AddNode([]string{"newkey"}, node.Content[0], nil)
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	if string(out) != initial {
		t.Errorf("expected node to be unchanged when adding nil, got:\n%s", string(out))
	}
}

func TestAddNodeComplexValue(t *testing.T) {
	testCases := []struct {
		name     string
		initial  string
		path     []string
		addNode  *yaml.Node
		expected string
	}{
		{
			name:    "add mapping node",
			initial: `{}`,
			path:    []string{"config"},
			addNode: &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "host"},
					{Kind: yaml.ScalarNode, Value: "localhost"},
					{Kind: yaml.ScalarNode, Value: "port"},
					{Kind: yaml.ScalarNode, Value: "8080"},
				},
			},
			expected: `config:
    host: localhost
    port: 8080
`,
		},
		{
			name:    "add sequence node",
			initial: `{}`,
			path:    []string{"items"},
			addNode: &yaml.Node{
				Kind: yaml.SequenceNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "item1"},
					{Kind: yaml.ScalarNode, Value: "item2"},
					{Kind: yaml.ScalarNode, Value: "item3"},
				},
			},
			expected: `items:
    - item1
    - item2
    - item3
`,
		},
		{
			name: "replace with complex node",
			initial: `simple: value
`,
			path: []string{"simple"},
			addNode: &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "nested"},
					{Kind: yaml.ScalarNode, Value: "value"},
				},
			},
			expected: `simple:
    nested: value
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tc.initial), &node)
			if err != nil {
				t.Fatalf("failed to unmarshal initial yaml: %v", err)
			}

			if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
				AddNode(tc.path, node.Content[0], tc.addNode)
			} else {
				t.Fatal("unexpected node structure")
			}

			out, err := yaml.Marshal(&node)
			if err != nil {
				t.Fatalf("failed to marshal result: %v", err)
			}

			// Unmarshal both to compare semantically rather than as strings
			var gotMap, expectedMap interface{}
			if err := yaml.Unmarshal(out, &gotMap); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if err := yaml.Unmarshal([]byte(tc.expected), &expectedMap); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			// Compare the semantic content
			gotYAML, _ := yaml.Marshal(gotMap)
			expectedYAML, _ := yaml.Marshal(expectedMap)
			if string(gotYAML) != string(expectedYAML) {
				t.Errorf("expected:\n%s\ngot:\n%s", string(expectedYAML), string(gotYAML))
			}
		})
	}
}

func TestAddNodeEmptyPath(t *testing.T) {
	initial := `key: value
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(initial), &node)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	addNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "newvalue",
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		AddNode([]string{}, node.Content[0], addNode)
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	if string(out) != initial {
		t.Errorf("expected node to be unchanged with empty path, got:\n%s", string(out))
	}
}

func TestAddNodePreservesExistingStructure(t *testing.T) {
	initial := `database:
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
`

	var node yaml.Node
	err := yaml.Unmarshal([]byte(initial), &node)
	if err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	addNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "newdb",
	}

	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		AddNode([]string{"database", "name"}, node.Content[0], addNode)
	}

	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(out, &result)
	if err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	db, ok := result["database"].(map[string]interface{})
	if !ok {
		t.Fatal("expected database to be a map")
	}

	if db["host"] != "localhost" {
		t.Errorf("expected host to be preserved as 'localhost', got %v", db["host"])
	}
	if db["port"] != 5432 {
		t.Errorf("expected port to be preserved as 5432, got %v", db["port"])
	}
	if db["name"] != "newdb" {
		t.Errorf("expected name to be 'newdb', got %v", db["name"])
	}

	servers, ok := result["servers"].([]interface{})
	if !ok {
		t.Fatal("expected servers to be preserved as a sequence")
	}
	if len(servers) != 2 {
		t.Errorf("expected 2 servers to be preserved, got %d", len(servers))
	}
}

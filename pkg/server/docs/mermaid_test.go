package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMermaidGenerator(t *testing.T) {
	gen := NewMermaidGenerator()

	// Test network data
	network := map[string]interface{}{
		"nodes": map[string]interface{}{
			"reader": map[string]interface{}{
				"type":   "FileReader",
				"status": "running",
			},
			"writer": map[string]interface{}{
				"type":   "FileWriter",
				"status": "waiting",
			},
		},
		"edges": []interface{}{
			map[string]interface{}{
				"from": "reader",
				"to":   "writer",
				"port": "data",
			},
		},
	}

	t.Run("generate diagram", func(t *testing.T) {
		gen.SetNetwork("test-flow", network)
		diagram, err := gen.GenerateDiagram("test-flow")
		require.NoError(t, err)
		assert.Contains(t, diagram, "reader[FileReader]:::running")
		assert.Contains(t, diagram, "writer[FileWriter]:::waiting")
		assert.Contains(t, diagram, "reader -->|data| writer")
	})
}

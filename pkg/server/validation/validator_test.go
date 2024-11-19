package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testValidator struct {
	allowedTypes map[string]bool
}

func newTestValidator() *testValidator {
	return &testValidator{
		allowedTypes: map[string]bool{
			"FileReader": true,
			"FileWriter": true,
			"Transform":  true,
		},
	}
}

func (v *testValidator) ValidateFlowConfig(config map[string]interface{}) error {
	// Test validation implementation
	return validateFlowConfig(config, v.allowedTypes)
}

func validateFlowConfig(config map[string]interface{}, allowedTypes map[string]bool) error {
	// Basic validation logic for testing
	if config == nil {
		return ErrEmptyConfig
	}

	id, ok := config["id"].(string)
	if !ok || id == "" {
		return ErrMissingID
	}

	nodes, ok := config["nodes"].(map[string]interface{})
	if !ok {
		return ErrInvalidNodes
	}

	for _, node := range nodes {
		nodeConfig, ok := node.(map[string]interface{})
		if !ok {
			return ErrInvalidNodeConfig
		}

		nodeType, ok := nodeConfig["type"].(string)
		if !ok || nodeType == "" {
			return ErrMissingNodeType
		}

		if !allowedTypes[nodeType] {
			return ErrInvalidNodeType
		}
	}

	return nil
}

func TestValidateFlowConfig(t *testing.T) {
	validator := newTestValidator()

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr error
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"id": "test-flow",
				"nodes": map[string]interface{}{
					"reader": map[string]interface{}{
						"type": "FileReader",
						"config": map[string]interface{}{
							"filename": "test.txt",
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: ErrEmptyConfig,
		},
		{
			name: "missing id",
			config: map[string]interface{}{
				"nodes": map[string]interface{}{},
			},
			wantErr: ErrMissingID,
		},
		{
			name: "invalid nodes",
			config: map[string]interface{}{
				"id":    "test-flow",
				"nodes": "invalid",
			},
			wantErr: ErrInvalidNodes,
		},
		{
			name: "invalid node config",
			config: map[string]interface{}{
				"id": "test-flow",
				"nodes": map[string]interface{}{
					"reader": "invalid",
				},
			},
			wantErr: ErrInvalidNodeConfig,
		},
		{
			name: "missing node type",
			config: map[string]interface{}{
				"id": "test-flow",
				"nodes": map[string]interface{}{
					"reader": map[string]interface{}{
						"config": map[string]interface{}{},
					},
				},
			},
			wantErr: ErrMissingNodeType,
		},
		{
			name: "invalid node type",
			config: map[string]interface{}{
				"id": "test-flow",
				"nodes": map[string]interface{}{
					"reader": map[string]interface{}{
						"type": "InvalidType",
					},
				},
			},
			wantErr: ErrInvalidNodeType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFlowConfig(tt.config)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

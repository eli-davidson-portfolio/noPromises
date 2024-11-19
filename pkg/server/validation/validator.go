package validation

// Validator defines the interface for flow configuration validation
type Validator interface {
	ValidateFlowConfig(config map[string]interface{}) error
}

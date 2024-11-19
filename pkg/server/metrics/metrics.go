package metrics

// Metrics defines the interface for metrics collection
type Metrics interface {
	RecordFlowCreation(flowID string)
	RecordFlowDeletion(flowID string)
	RecordFlowStart(flowID string)
	RecordFlowStop(flowID string)
}

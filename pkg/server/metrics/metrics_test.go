package metrics

import (
	"testing"
)

type mockMetrics struct {
	flowCreations int
	flowDeletions int
	flowStarts    int
	flowStops     int
}

func (m *mockMetrics) RecordFlowCreation(_ string) {
	m.flowCreations++
}

func (m *mockMetrics) RecordFlowDeletion(_ string) {
	m.flowDeletions++
}

func (m *mockMetrics) RecordFlowStart(_ string) {
	m.flowStarts++
}

func (m *mockMetrics) RecordFlowStop(_ string) {
	m.flowStops++
}

func TestMetricsRecording(t *testing.T) {
	m := &mockMetrics{}

	m.RecordFlowCreation("test-flow")
	m.RecordFlowStart("test-flow")
	m.RecordFlowStop("test-flow")
	m.RecordFlowDeletion("test-flow")

	if m.flowCreations != 1 {
		t.Error("Flow creation not recorded")
	}
	if m.flowStarts != 1 {
		t.Error("Flow start not recorded")
	}
	if m.flowStops != 1 {
		t.Error("Flow stop not recorded")
	}
	if m.flowDeletions != 1 {
		t.Error("Flow deletion not recorded")
	}
}

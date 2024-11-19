package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFlowListEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "get flow list",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   `<div class="flow-item">test-flow-1</div>`,
		},
		{
			name:           "create new flow",
			method:         http.MethodPost,
			body:           `{"id": "new-flow", "config": {"type": "test"}}`,
			expectedStatus: http.StatusOK,
			expectedBody:   `<div class="flow-item">new-flow</div>`,
		},
		{
			name:           "invalid method",
			method:         http.MethodPut,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer()
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, "/api/v1/flows", bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/api/v1/flows", nil)
			}
			w := httptest.NewRecorder()

			server.HandleFlows()(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedBody != "" && !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestFlowVisualization(t *testing.T) {
	tests := []struct {
		name           string
		flowID         string
		expectedStatus int
		expectedNodes  int
	}{
		{
			name:           "valid flow visualization",
			flowID:         "test-flow-1",
			expectedStatus: http.StatusOK,
			expectedNodes:  2,
		},
		{
			name:           "flow not found",
			flowID:         "non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupTestServer()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/flows/"+tt.flowID+"/viz", nil)
			w := httptest.NewRecorder()

			server.HandleFlowVisualization()(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Nodes []map[string]string `json:"nodes"`
					Edges []map[string]string `json:"edges"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(response.Nodes) != tt.expectedNodes {
					t.Errorf("expected %d nodes, got %d", tt.expectedNodes, len(response.Nodes))
				}
			}
		})
	}
}

package web

import (
	"html/template"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockFlowManager struct {
	flows []ManagedFlow
}

func (m *mockFlowManager) List() []ManagedFlow {
	return m.flows
}

func (m *mockFlowManager) Get(id string) (*ManagedFlow, bool) {
	for _, flow := range m.flows {
		if flow.ID == id {
			return &flow, true
		}
	}
	return nil, false
}

func setupTestServer() *Server {
	// Create test template with simpler HTML that matches test expectations
	tmpl := template.Must(template.New("index.html").Parse(`
		<html><body>
			<h1>{{.Title}}</h1>
			<div>
				{{range .Flows}}
					<div>{{.ID}}: {{.State}}</div>
				{{end}}
			</div>
		</body></html>
	`))

	// Create mock flow manager
	mockFM := &mockFlowManager{
		flows: []ManagedFlow{
			{ID: "test-flow-1", State: "running"},
			{ID: "test-flow-2", State: "stopped"},
		},
	}

	return NewServer(
		WithFlowManager(mockFM),
		WithTemplates(tmpl),
	)
}

func TestHandleHome(t *testing.T) {
	srv := setupTestServer()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	srv.HandleHome()(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "NoPromises Flow Manager")
}

func TestHandleFlows(t *testing.T) {
	srv := setupTestServer()

	t.Run("list flows", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/flows", nil)
		w := httptest.NewRecorder()

		srv.HandleFlows()(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "test-flow-1")
		assert.Contains(t, w.Body.String(), "test-flow-2")
	})
}

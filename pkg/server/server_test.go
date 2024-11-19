package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer encapsulates server testing utilities
type TestServer struct {
	*Server
	t *testing.T
}

// MockFileReader implements Process interface for testing
type MockFileReader struct{}

func (m *MockFileReader) Start(_ context.Context) error { return nil }
func (m *MockFileReader) Stop(_ context.Context) error  { return nil }

// MockFileReaderFactory implements ProcessFactory interface
type MockFileReaderFactory struct{}

func (f *MockFileReaderFactory) Create(_ map[string]interface{}) (Process, error) {
	return &MockFileReader{}, nil
}

func NewTestServer(t *testing.T) *TestServer {
	server, err := NewServer(Config{
		Port: 0, // Random port for testing
	})
	require.NoError(t, err)

	// Register FileReader process type for testing
	server.RegisterProcessType("FileReader", &MockFileReaderFactory{})

	return &TestServer{
		Server: server,
		t:      t,
	}
}

// Helper method to make test requests
func (ts *TestServer) request(method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader *bytes.Buffer
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(ts.t, err)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil) // Initialize empty buffer for nil body
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	ts.Server.Handler.ServeHTTP(rr, req)
	return rr
}

func TestCreateFlow(t *testing.T) {
	ts := NewTestServer(t)

	tests := []struct {
		name       string
		flowConfig map[string]interface{}
		wantStatus int
		wantError  bool
	}{
		{
			name: "valid flow",
			flowConfig: map[string]interface{}{
				"id": "test-flow",
				"nodes": map[string]interface{}{
					"reader": map[string]interface{}{
						"type": "FileReader",
						"config": map[string]interface{}{
							"filename": "test.txt",
						},
					},
				},
				"edges": []interface{}{},
			},
			wantStatus: http.StatusCreated,
			wantError:  false,
		},
		{
			name: "missing flow id",
			flowConfig: map[string]interface{}{
				"nodes": map[string]interface{}{},
				"edges": []interface{}{},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "invalid node type",
			flowConfig: map[string]interface{}{
				"id": "test-flow",
				"nodes": map[string]interface{}{
					"reader": map[string]interface{}{
						"type": "InvalidType",
					},
				},
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := ts.request(http.MethodPost, "/api/v1/flows", tt.flowConfig)

			assert.Equal(t, tt.wantStatus, rr.Code)

			var resp map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, resp, "error")
			} else {
				assert.Contains(t, resp, "data")
			}
		})
	}
}

func TestGetFlow(t *testing.T) {
	ts := NewTestServer(t)

	// Create a test flow first
	flowConfig := map[string]interface{}{
		"id": "test-flow",
		"nodes": map[string]interface{}{
			"reader": map[string]interface{}{
				"type": "FileReader",
				"config": map[string]interface{}{
					"filename": "test.txt",
				},
			},
		},
		"edges": []interface{}{},
	}

	createResp := ts.request(http.MethodPost, "/api/v1/flows", flowConfig)
	require.Equal(t, http.StatusCreated, createResp.Code)

	tests := []struct {
		name       string
		flowID     string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "existing flow",
			flowID:     "test-flow",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "non-existent flow",
			flowID:     "missing-flow",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := ts.request(http.MethodGet, "/api/v1/flows/"+tt.flowID, nil)

			assert.Equal(t, tt.wantStatus, rr.Code)

			var resp map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, resp, "error")
			} else {
				assert.Contains(t, resp, "data")
			}
		})
	}
}

func TestStartFlow(t *testing.T) {
	ts := NewTestServer(t)

	// Create a test flow
	flowConfig := map[string]interface{}{
		"id": "test-flow",
		"nodes": map[string]interface{}{
			"reader": map[string]interface{}{
				"type": "FileReader",
				"config": map[string]interface{}{
					"filename": "test.txt",
				},
			},
		},
		"edges": []interface{}{},
	}

	createResp := ts.request(http.MethodPost, "/api/v1/flows", flowConfig)
	require.Equal(t, http.StatusCreated, createResp.Code)

	tests := []struct {
		name       string
		flowID     string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "start existing flow",
			flowID:     "test-flow",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "start non-existent flow",
			flowID:     "missing-flow",
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := ts.request(http.MethodPost, "/api/v1/flows/"+tt.flowID+"/start", nil)

			assert.Equal(t, tt.wantStatus, rr.Code)

			var resp map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&resp)
			require.NoError(t, err)

			if tt.wantError {
				assert.Contains(t, resp, "error")
			} else {
				assert.Contains(t, resp, "data")
			}
		})
	}
}

func TestFlowLifecycle(t *testing.T) {
	ts := NewTestServer(t)

	// 1. Create flow
	flowConfig := map[string]interface{}{
		"id": "lifecycle-flow",
		"nodes": map[string]interface{}{
			"reader": map[string]interface{}{
				"type": "FileReader",
				"config": map[string]interface{}{
					"filename": "test.txt",
				},
			},
		},
		"edges": []interface{}{},
	}

	createResp := ts.request(http.MethodPost, "/api/v1/flows", flowConfig)
	require.Equal(t, http.StatusCreated, createResp.Code)

	// 2. Start flow
	startResp := ts.request(http.MethodPost, "/api/v1/flows/lifecycle-flow/start", nil)
	require.Equal(t, http.StatusOK, startResp.Code)

	// 3. Check status
	time.Sleep(100 * time.Millisecond) // Allow flow to start
	statusResp := ts.request(http.MethodGet, "/api/v1/flows/lifecycle-flow/status", nil)
	require.Equal(t, http.StatusOK, statusResp.Code)

	var status map[string]interface{}
	err := json.NewDecoder(statusResp.Body).Decode(&status)
	require.NoError(t, err)
	assert.Equal(t, "running", status["data"].(map[string]interface{})["state"])

	// 4. Stop flow
	stopResp := ts.request(http.MethodPost, "/api/v1/flows/lifecycle-flow/stop", nil)
	require.Equal(t, http.StatusOK, stopResp.Code)

	// 5. Delete flow
	deleteResp := ts.request(http.MethodDelete, "/api/v1/flows/lifecycle-flow", nil)
	require.Equal(t, http.StatusNoContent, deleteResp.Code)

	// 6. Verify deletion
	getResp := ts.request(http.MethodGet, "/api/v1/flows/lifecycle-flow", nil)
	require.Equal(t, http.StatusNotFound, getResp.Code)
}

func TestConcurrentFlowOperations(t *testing.T) {
	ts := NewTestServer(t)

	const numFlows = 10
	errCh := make(chan error, numFlows*2)

	// Create and start multiple flows concurrently
	for i := 0; i < numFlows; i++ {
		go func(id int) {
			flowConfig := map[string]interface{}{
				"id": fmt.Sprintf("concurrent-flow-%d", id),
				"nodes": map[string]interface{}{
					"reader": map[string]interface{}{
						"type": "FileReader",
						"config": map[string]interface{}{
							"filename": "test.txt",
						},
					},
				},
				"edges": []interface{}{},
			}

			// Create flow
			createResp := ts.request(http.MethodPost, "/api/v1/flows", flowConfig)
			if createResp.Code != http.StatusCreated {
				errCh <- fmt.Errorf("failed to create flow %d: %d", id, createResp.Code)
				return
			}

			// Start flow
			startResp := ts.request(http.MethodPost, fmt.Sprintf("/api/v1/flows/concurrent-flow-%d/start", id), nil)
			if startResp.Code != http.StatusOK {
				errCh <- fmt.Errorf("failed to start flow %d: %d", id, startResp.Code)
				return
			}

			errCh <- nil
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numFlows; i++ {
		select {
		case err := <-errCh:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent operations")
		}
	}

	// Wait for flows to transition to running state
	time.Sleep(100 * time.Millisecond)

	// Verify all flows are running
	for i := 0; i < numFlows; i++ {
		statusResp := ts.request(http.MethodGet, fmt.Sprintf("/api/v1/flows/concurrent-flow-%d/status", i), nil)
		require.Equal(t, http.StatusOK, statusResp.Code)

		var status map[string]interface{}
		err := json.NewDecoder(statusResp.Body).Decode(&status)
		require.NoError(t, err)
		assert.Equal(t, "running", status["data"].(map[string]interface{})["state"])
	}
}

func TestServerWithDocs(t *testing.T) {
	// Create temporary docs directory
	tmpDir := t.TempDir()

	// Create test documentation file
	testDoc := filepath.Join(tmpDir, "test.md")
	err := os.WriteFile(testDoc, []byte("# Test Documentation"), 0644)
	require.NoError(t, err)

	// Create server with docs configuration
	srv, err := NewServer(Config{
		Port:     8080,
		DocsPath: tmpDir,
	})
	require.NoError(t, err)
	require.NotNil(t, srv.docsServer)

	// Create test server
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "serve markdown file",
			path:           "/docs/test.md",
			expectedStatus: http.StatusOK,
			expectedBody:   "# Test Documentation",
		},
		{
			name:           "handle missing file",
			path:           "/docs/missing.md",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "block directory traversal",
			path:           "/docs/../private.txt",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.path)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}

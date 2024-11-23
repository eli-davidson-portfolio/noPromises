package docs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// MermaidGenerator handles network diagram generation
type MermaidGenerator struct {
	networks sync.Map
}

// NewMermaidGenerator creates a new diagram generator
func NewMermaidGenerator() *MermaidGenerator {
	return &MermaidGenerator{}
}

// SetNetwork updates the network definition
func (m *MermaidGenerator) SetNetwork(id string, network map[string]interface{}) {
	m.networks.Store(id, network)
}

// GetNetwork retrieves a network definition
func (m *MermaidGenerator) GetNetwork(id string) (map[string]interface{}, bool) {
	if val, ok := m.networks.Load(id); ok {
		return val.(map[string]interface{}), true
	}
	return nil, false
}

// GenerateDiagram creates a Mermaid diagram for a network
func (m *MermaidGenerator) GenerateDiagram(id string) (string, error) {
	network, ok := m.GetNetwork(id)
	if !ok {
		return "", fmt.Errorf("network not found: %s", id)
	}

	// Basic diagram generation
	diagram := "graph LR\n"

	// Add nodes
	if nodes, ok := network["nodes"].(map[string]interface{}); ok {
		for id, node := range nodes {
			if nodeData, ok := node.(map[string]interface{}); ok {
				nodeType := nodeData["type"].(string)
				status := nodeData["status"].(string)
				diagram += fmt.Sprintf("    %s[%s]:::%s\n", id, nodeType, status)
			}
		}
	}

	// Add edges
	if edges, ok := network["edges"].([]interface{}); ok {
		for _, e := range edges {
			if edge, ok := e.(map[string]interface{}); ok {
				from := edge["from"].(string)
				to := edge["to"].(string)
				port := edge["port"].(string)
				diagram += fmt.Sprintf("    %s -->|%s| %s\n", from, port, to)
			}
		}
	}

	return diagram, nil
}

// HandleNetworkDiagram serves network diagrams
func (m *MermaidGenerator) HandleNetworkDiagram(w http.ResponseWriter, _ *http.Request, id string) {
	diagram, err := m.GenerateDiagram(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"diagram": diagram,
	}); err != nil {
		log.Printf("[ERROR] Failed to encode diagram response: %v", err)
	}
}

package docs

import (
	"fmt"
	"strings"
)

// MermaidGenerator generates Mermaid diagrams from network configurations
type MermaidGenerator struct {
	networks map[string]interface{}
}

// NewMermaidGenerator creates a new MermaidGenerator instance
func NewMermaidGenerator() *MermaidGenerator {
	return &MermaidGenerator{
		networks: make(map[string]interface{}),
	}
}

// SetNetwork updates or adds a network configuration
func (g *MermaidGenerator) SetNetwork(id string, network interface{}) {
	g.networks[id] = network
}

// GenerateFlowDiagram creates a Mermaid diagram from a network configuration
func (g *MermaidGenerator) GenerateFlowDiagram(networkID string) (string, error) {
	network, exists := g.networks[networkID]
	if !exists {
		return "", fmt.Errorf("network not found: %s", networkID)
	}

	var diagram strings.Builder
	diagram.WriteString("graph LR\n")

	// Convert network interface to map
	netMap, ok := network.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid network configuration")
	}

	// Add nodes
	if nodes, ok := netMap["nodes"].(map[string]interface{}); ok {
		for id, node := range nodes {
			nodeMap, ok := node.(map[string]interface{})
			if !ok {
				continue
			}
			nodeType := nodeMap["type"].(string)
			status := nodeMap["status"].(string)
			diagram.WriteString(fmt.Sprintf("    %s[%s]:::%s\n", id, nodeType, status))
		}
	}

	// Add edges
	if edges, ok := netMap["edges"].([]interface{}); ok {
		for _, edge := range edges {
			edgeMap, ok := edge.(map[string]interface{})
			if !ok {
				continue
			}
			from := edgeMap["from"].(string)
			to := edgeMap["to"].(string)
			port := edgeMap["port"].(string)
			diagram.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", from, port, to))
		}
	}

	// Add style definitions
	diagram.WriteString("\n    classDef running fill:#d4edda,stroke:#28a745;\n")
	diagram.WriteString("    classDef waiting fill:#fff3cd,stroke:#ffc107;\n")
	diagram.WriteString("    classDef error fill:#f8d7da,stroke:#dc3545;\n")

	return diagram.String(), nil
}

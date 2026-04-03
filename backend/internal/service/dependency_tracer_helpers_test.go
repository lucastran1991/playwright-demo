package service

import (
	"testing"

	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
)

// TestFilterByTypes_EmptyAllowedSet tests filtering with empty allowed set.
func TestFilterByTypes_EmptyAllowedSet(t *testing.T) {
	nodes := []repository.TracedNode{
		{ID: 1, NodeID: "node-1", Name: "Node 1", NodeType: "Rack", Level: 1},
		{ID: 2, NodeID: "node-2", Name: "Node 2", NodeType: "RPP", Level: 2},
	}

	// Empty allowed set should return all nodes
	filtered := filterByTypes(nodes, make(map[string]bool))
	if len(filtered) != 2 {
		t.Errorf("Expected 2 nodes with empty allowed set, got %d", len(filtered))
	}
}

// TestFilterByTypes_AllFiltered tests filtering that removes all nodes.
func TestFilterByTypes_AllFiltered(t *testing.T) {
	nodes := []repository.TracedNode{
		{ID: 1, NodeID: "node-1", Name: "Node 1", NodeType: "Rack", Level: 1},
		{ID: 2, NodeID: "node-2", Name: "Node 2", NodeType: "Row", Level: 2},
	}

	allowed := map[string]bool{
		"RPP": true,
		"UPS": true,
	}

	filtered := filterByTypes(nodes, allowed)
	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered nodes, got %d", len(filtered))
	}
}

// TestFilterByTypes_PartialFilter tests filtering that allows some nodes.
func TestFilterByTypes_PartialFilter(t *testing.T) {
	nodes := []repository.TracedNode{
		{ID: 1, NodeID: "node-1", Name: "Node 1", NodeType: "Rack", Level: 1},
		{ID: 2, NodeID: "node-2", Name: "Node 2", NodeType: "RPP", Level: 1},
		{ID: 3, NodeID: "node-3", Name: "Node 3", NodeType: "UPS", Level: 2},
		{ID: 4, NodeID: "node-4", Name: "Node 4", NodeType: "Row", Level: 2},
	}

	allowed := map[string]bool{
		"RPP": true,
		"UPS": true,
	}

	filtered := filterByTypes(nodes, allowed)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered nodes, got %d", len(filtered))
	}

	// Verify correct nodes are kept
	for _, n := range filtered {
		if n.NodeType != "RPP" && n.NodeType != "UPS" {
			t.Errorf("Unexpected NodeType %q in filtered results", n.NodeType)
		}
	}
}

// TestFilterByTypes_EmptyNodesList tests filtering with empty node list.
func TestFilterByTypes_EmptyNodesList(t *testing.T) {
	var nodes []repository.TracedNode

	allowed := map[string]bool{
		"RPP": true,
		"UPS": true,
	}

	filtered := filterByTypes(nodes, allowed)
	if len(filtered) != 0 {
		t.Errorf("Expected 0 filtered nodes from empty list, got %d", len(filtered))
	}
}

// TestGroupByLevel_SingleLevel tests grouping at single level.
func TestGroupByLevel_SingleLevel(t *testing.T) {
	nodes := []repository.TracedNode{
		{ID: 1, NodeID: "node-1", Name: "Node 1", NodeType: "RPP", Level: 1},
		{ID: 2, NodeID: "node-2", Name: "Node 2", NodeType: "UPS", Level: 1},
	}

	groups := groupByLevel(nodes, "Electrical System")
	if len(groups) != 1 {
		t.Errorf("Expected 1 level group, got %d", len(groups))
	}

	if groups[0].Level != 1 {
		t.Errorf("Expected level 1, got %d", groups[0].Level)
	}
	if groups[0].Topology != "Electrical System" {
		t.Errorf("Expected topology 'Electrical System', got '%s'", groups[0].Topology)
	}
	if len(groups[0].Nodes) != 2 {
		t.Errorf("Expected 2 nodes at level 1, got %d", len(groups[0].Nodes))
	}
}

// TestGroupByLevel_MultipleLevels tests grouping nodes across multiple levels.
func TestGroupByLevel_MultipleLevels(t *testing.T) {
	nodes := []repository.TracedNode{
		{ID: 1, NodeID: "node-1", Name: "Node 1", NodeType: "RPP", Level: 1},
		{ID: 2, NodeID: "node-2", Name: "Node 2", NodeType: "UPS", Level: 1},
		{ID: 3, NodeID: "node-3", Name: "Node 3", NodeType: "Generator", Level: 2},
		{ID: 4, NodeID: "node-4", Name: "Node 4", NodeType: "Utility Feed", Level: 3},
	}

	groups := groupByLevel(nodes, "Electrical System")
	if len(groups) != 3 {
		t.Errorf("Expected 3 level groups, got %d", len(groups))
	}

	// Verify levels are present
	levelMap := make(map[int]bool)
	for _, g := range groups {
		levelMap[g.Level] = true
		if g.Topology != "Electrical System" {
			t.Errorf("Expected topology 'Electrical System', got '%s'", g.Topology)
		}
	}

	if !levelMap[1] || !levelMap[2] || !levelMap[3] {
		t.Errorf("Not all levels present: %v", levelMap)
	}
}

// TestGroupByLevel_EmptyNodes tests grouping with empty node list.
func TestGroupByLevel_EmptyNodes(t *testing.T) {
	var nodes []repository.TracedNode

	groups := groupByLevel(nodes, "Electrical System")
	if len(groups) != 0 {
		t.Errorf("Expected 0 level groups from empty list, got %d", len(groups))
	}
}

// TestGroupByLevel_PreservesNodeData tests that node data is preserved during grouping.
func TestGroupByLevel_PreservesNodeData(t *testing.T) {
	nodes := []repository.TracedNode{
		{ID: 42, NodeID: "rpp-001", Name: "RPP Room 001", NodeType: "RPP", Level: 1},
	}

	groups := groupByLevel(nodes, "Electrical System")
	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	if len(groups[0].Nodes) != 1 {
		t.Fatalf("Expected 1 node in group, got %d", len(groups[0].Nodes))
	}

	node := groups[0].Nodes[0]
	if node.ID != 42 {
		t.Errorf("Expected node ID 42, got %d", node.ID)
	}
	if node.NodeID != "rpp-001" {
		t.Errorf("Expected node_id 'rpp-001', got '%s'", node.NodeID)
	}
	if node.Name != "RPP Room 001" {
		t.Errorf("Expected name 'RPP Room 001', got '%s'", node.Name)
	}
	if node.NodeType != "RPP" {
		t.Errorf("Expected node_type 'RPP', got '%s'", node.NodeType)
	}
}

// TestGroupDepRules_UpstreamAndLocal tests grouping of dependency rules with mocked tracer.
func TestGroupDepRules_UpstreamAndLocal(t *testing.T) {
	// Create a DependencyTracer with empty lookups for testing
	tracer := &DependencyTracer{
		topoLookup: map[string]string{
			"RPP":       "Electrical System",
			"RDHx":      "Cooling System",
			"Air Zone":  "Cooling System",
		},
		slugLookup: make(map[string]string),
	}

	rules := []model.DependencyRule{
		{
			NodeType:                "Rack",
			DependencyNodeType:      "RPP",
			TopologicalRelationship: "Upstream",
			UpstreamLevel:           intPtr(1),
		},
		{
			NodeType:                "Rack",
			DependencyNodeType:      "RDHx",
			TopologicalRelationship: "Local",
			UpstreamLevel:           nil,
		},
		{
			NodeType:                "Rack",
			DependencyNodeType:      "Air Zone",
			TopologicalRelationship: "Upstream",
			UpstreamLevel:           intPtr(1),
		},
	}

	upstream, local := tracer.groupDepRules(rules)

	// Check upstream grouping
	if len(upstream) != 2 {
		t.Errorf("Expected 2 upstream topologies, got %d", len(upstream))
	}

	// Electrical System should have RPP
	if !upstream["Electrical System"]["RPP"] {
		t.Error("Expected RPP in Electrical System upstream")
	}

	// Cooling System should have Air Zone
	if !upstream["Cooling System"]["Air Zone"] {
		t.Error("Expected Air Zone in Cooling System upstream")
	}

	// Check local grouping
	if len(local) != 1 {
		t.Errorf("Expected 1 local topology, got %d", len(local))
	}

	if !local["Cooling System"]["RDHx"] {
		t.Error("Expected RDHx in Cooling System local")
	}
}

// TestGroupImpactRules_DownstreamAndLoad tests grouping of impact rules.
func TestGroupImpactRules_DownstreamAndLoad(t *testing.T) {
	// Create a DependencyTracer with mocked lookups
	tracer := &DependencyTracer{
		topoLookup: map[string]string{
			"Rack PDU":       "Electrical System",
			"Rack":           "Spatial Topology",
			"Room Bundle":    "Whitespace Blueprint",
		},
		slugLookup: make(map[string]string),
	}

	rules := []model.ImpactRule{
		{
			NodeType:                "RPP",
			ImpactNodeType:          "Rack PDU",
			TopologicalRelationship: "Downstream",
			DownstreamLevel:         intPtr(1),
		},
		{
			NodeType:                "RPP",
			ImpactNodeType:          "Rack",
			TopologicalRelationship: "Load",
			DownstreamLevel:         nil,
		},
		{
			NodeType:                "UPS",
			ImpactNodeType:          "Room Bundle",
			TopologicalRelationship: "Load",
			DownstreamLevel:         nil,
		},
	}

	downstream, load := tracer.groupImpactRules(rules)

	// Check downstream grouping
	if len(downstream) != 1 {
		t.Errorf("Expected 1 downstream topology, got %d", len(downstream))
	}

	if !downstream["Electrical System"]["Rack PDU"] {
		t.Error("Expected Rack PDU in Electrical System downstream")
	}

	// Check load grouping
	if len(load) != 2 {
		t.Errorf("Expected 2 load topologies, got %d", len(load))
	}

	if !load["Spatial Topology"]["Rack"] {
		t.Error("Expected Rack in Spatial Topology load")
	}

	if !load["Whitespace Blueprint"]["Room Bundle"] {
		t.Error("Expected Room Bundle in Whitespace Blueprint load")
	}
}

// Helper function to create int pointers for tests
func intPtr(i int) *int {
	return &i
}

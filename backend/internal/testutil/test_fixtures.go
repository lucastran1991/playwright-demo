package testutil

import (
	"testing"

	"github.com/user/app/internal/model"
	"gorm.io/gorm"
)

// SeedTraceFixtures inserts a minimal topology graph that exercises all trace
// paths (upstream, downstream, local, load, bridge). Returns node_id → DB ID map.
func SeedTraceFixtures(t *testing.T, db *gorm.DB) map[string]uint {
	t.Helper()

	// 1. Blueprint types (4 topologies)
	types := []model.BlueprintType{
		{Name: "Electrical System", Slug: "electrical-system", FolderName: "Electrical System"},
		{Name: "Cooling System", Slug: "cooling-system", FolderName: "Cooling System"},
		{Name: "Spatial Topology", Slug: "spatial-topology", FolderName: "Spatial Topology"},
		{Name: "Whitespace", Slug: "whitespace", FolderName: "Whitespace"},
	}
	for i := range types {
		if err := db.Create(&types[i]).Error; err != nil {
			t.Fatalf("seed blueprint_type %q: %v", types[i].Name, err)
		}
	}
	typeID := make(map[string]uint)
	for _, bt := range types {
		typeID[bt.Slug] = bt.ID
	}

	// 2. Blueprint nodes (10 nodes spanning all topologies)
	nodes := []model.BlueprintNode{
		{NodeID: "GEN-01", Name: "Generator 1", NodeType: "Generator"},
		{NodeID: "UPS-01", Name: "UPS Unit 1", NodeType: "UPS"},
		{NodeID: "RPP-01", Name: "RPP Panel 1", NodeType: "RPP"},
		{NodeID: "RACKPDU-01", Name: "Rack PDU 1", NodeType: "Rack PDU"},
		{NodeID: "RACK-01", Name: "Rack 1", NodeType: "Rack"},
		{NodeID: "RDHX-01", Name: "RDHx Unit 1", NodeType: "RDHx"},
		{NodeID: "AIRZONE-01", Name: "Air Zone 1", NodeType: "Air Zone"},
		{NodeID: "ROW-01", Name: "Row 1", NodeType: "Row"},
		{NodeID: "ZONE-01", Name: "Zone 1", NodeType: "Zone"},
		{NodeID: "CC-01", Name: "Capacity Cell 1", NodeType: "Capacity Cell"},
	}
	for i := range nodes {
		if err := db.Create(&nodes[i]).Error; err != nil {
			t.Fatalf("seed node %q: %v", nodes[i].NodeID, err)
		}
	}
	nodeMap := make(map[string]uint, len(nodes))
	for _, n := range nodes {
		nodeMap[n.NodeID] = n.ID
	}

	// 3. Blueprint edges — topology-specific parent→child relationships
	edges := []model.BlueprintEdge{
		// Electrical chain: GEN → UPS → RPP → RACKPDU
		{BlueprintTypeID: typeID["electrical-system"], FromNodeID: nodeMap["GEN-01"], ToNodeID: nodeMap["UPS-01"]},
		{BlueprintTypeID: typeID["electrical-system"], FromNodeID: nodeMap["UPS-01"], ToNodeID: nodeMap["RPP-01"]},
		{BlueprintTypeID: typeID["electrical-system"], FromNodeID: nodeMap["RPP-01"], ToNodeID: nodeMap["RACKPDU-01"]},
		// Cooling: RDHX ↔ AIRZONE (peer/local)
		{BlueprintTypeID: typeID["cooling-system"], FromNodeID: nodeMap["RDHX-01"], ToNodeID: nodeMap["AIRZONE-01"]},
		// Spatial: ZONE → ROW → RACK → RACKPDU
		{BlueprintTypeID: typeID["spatial-topology"], FromNodeID: nodeMap["ZONE-01"], ToNodeID: nodeMap["ROW-01"]},
		{BlueprintTypeID: typeID["spatial-topology"], FromNodeID: nodeMap["ROW-01"], ToNodeID: nodeMap["RACK-01"]},
		{BlueprintTypeID: typeID["spatial-topology"], FromNodeID: nodeMap["RACK-01"], ToNodeID: nodeMap["RACKPDU-01"]},
		// Whitespace: CC → RACK (bridge path for CC tracing)
		{BlueprintTypeID: typeID["whitespace"], FromNodeID: nodeMap["CC-01"], ToNodeID: nodeMap["RACK-01"]},
	}
	for i := range edges {
		if err := db.Create(&edges[i]).Error; err != nil {
			t.Fatalf("seed edge %d: %v", i, err)
		}
	}

	// 4. Capacity node types (type-level metadata)
	capTypes := []model.CapacityNodeType{
		{NodeType: "Generator", Topology: "Electrical System", IsCapacityNode: false},
		{NodeType: "UPS", Topology: "Electrical System", IsCapacityNode: true},
		{NodeType: "RPP", Topology: "Electrical System", IsCapacityNode: true},
		{NodeType: "Rack PDU", Topology: "Electrical System", IsCapacityNode: false},
		{NodeType: "RDHx", Topology: "Cooling System", IsCapacityNode: true},
		{NodeType: "Air Zone", Topology: "Cooling System", IsCapacityNode: false},
		{NodeType: "Rack", Topology: "Spatial Topology", IsCapacityNode: false},
		{NodeType: "Row", Topology: "Spatial Topology", IsCapacityNode: false},
		{NodeType: "Zone", Topology: "Spatial Topology", IsCapacityNode: false},
		{NodeType: "Capacity Cell", Topology: "Whitespace", IsCapacityNode: false},
	}
	for i := range capTypes {
		if err := db.Create(&capTypes[i]).Error; err != nil {
			t.Fatalf("seed cap_type %q: %v", capTypes[i].NodeType, err)
		}
	}

	// 5. Dependency rules (what Rack PDU depends on)
	depRules := []model.DependencyRule{
		{NodeType: "Rack PDU", DependencyNodeType: "RPP", RelationshipType: "Dependency",
			TopologicalRelationship: "Upstream", UpstreamLevel: intPtr(1)},
		{NodeType: "Rack PDU", DependencyNodeType: "UPS", RelationshipType: "Dependency",
			TopologicalRelationship: "Upstream", UpstreamLevel: intPtr(2)},
		{NodeType: "Rack PDU", DependencyNodeType: "RDHx", RelationshipType: "Dependency",
			TopologicalRelationship: "Local", UpstreamLevel: nil},
		// RPP depends on UPS upstream (enables RPP upstream tests)
		{NodeType: "RPP", DependencyNodeType: "UPS", RelationshipType: "Dependency",
			TopologicalRelationship: "Upstream", UpstreamLevel: intPtr(1)},
	}
	for i := range depRules {
		if err := db.Create(&depRules[i]).Error; err != nil {
			t.Fatalf("seed dep_rule %d: %v", i, err)
		}
	}

	// 6. Impact rules (what RPP impacts)
	impRules := []model.ImpactRule{
		{NodeType: "RPP", ImpactNodeType: "Rack PDU", TopologicalRelationship: "Downstream", DownstreamLevel: intPtr(1)},
		{NodeType: "RPP", ImpactNodeType: "Rack", TopologicalRelationship: "Load", DownstreamLevel: nil},
	}
	for i := range impRules {
		if err := db.Create(&impRules[i]).Error; err != nil {
			t.Fatalf("seed impact_rule %d: %v", i, err)
		}
	}

	return nodeMap
}

func intPtr(i int) *int { return &i }

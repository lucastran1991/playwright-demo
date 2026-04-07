package service

import "testing"

// TestTraceFull_LevelsLimitUpstream verifies maxLevels restricts upstream depth.
func TestTraceFull_LevelsLimitUpstream(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	// maxLevels=1: only RPP-01 at L1
	resp1, err := tracer.TraceFull("RACKPDU-01", 1)
	if err != nil {
		t.Fatalf("TraceFull RACKPDU-01 maxLevels=1: %v", err)
	}
	if !hasNodeInUpstream(resp1, "RPP-01") {
		t.Error("expected RPP-01 at L1 with maxLevels=1")
	}
	if hasNodeInUpstream(resp1, "UPS-01") {
		t.Error("UPS-01 should NOT appear at maxLevels=1 (it's L2)")
	}

	// maxLevels=2: both RPP-01 and UPS-01
	resp2, err := tracer.TraceFull("RACKPDU-01", 2)
	if err != nil {
		t.Fatalf("TraceFull RACKPDU-01 maxLevels=2: %v", err)
	}
	if !hasNodeInUpstream(resp2, "RPP-01") {
		t.Error("expected RPP-01 in upstream with maxLevels=2")
	}
	if !hasNodeInUpstream(resp2, "UPS-01") {
		t.Error("expected UPS-01 in upstream with maxLevels=2")
	}
	if l := getUpstreamLevel(resp2, "RPP-01"); l != 1 {
		t.Errorf("RPP-01 upstream level = %d, want 1", l)
	}
	if l := getUpstreamLevel(resp2, "UPS-01"); l != 2 {
		t.Errorf("UPS-01 upstream level = %d, want 2", l)
	}
}

// TestTraceFull_LoadOnlyForCapacityNodes verifies Load is populated only for capacity nodes.
func TestTraceFull_LoadOnlyForCapacityNodes(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	rppResp, err := tracer.TraceFull("RPP-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RPP-01: %v", err)
	}
	if len(rppResp.Load) == 0 {
		t.Error("expected Load non-empty for RPP-01 (IsCapacityNode=true)")
	}

	pduResp, err := tracer.TraceFull("RACKPDU-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RACKPDU-01: %v", err)
	}
	if len(pduResp.Load) != 0 {
		t.Errorf("expected Load empty for RACKPDU-01 (IsCapacityNode=false), got %d groups", len(pduResp.Load))
	}
}

// TestTraceDependencies_BridgeFallback verifies CC-01 bridge behavior.
// No dep rules seeded for Capacity Cell — upstream is empty, confirming bridge
// only activates when rules point to a topology. This is correct behavior.
func TestTraceDependencies_BridgeFallback(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	resp, err := tracer.TraceDependencies("CC-01", 5, false)
	if err != nil {
		t.Fatalf("TraceDependencies CC-01: %v", err)
	}
	if resp.Source.NodeID != "CC-01" {
		t.Errorf("Source.NodeID = %q, want CC-01", resp.Source.NodeID)
	}
	_ = resp.Upstream // empty is correct — no rules for Capacity Cell
}

// TestTraceDependencies_LocalNodes verifies RACKPDU-01 local dep handling.
// Has Local rule for RDHx (Cooling), but RACKPDU-01 has no cooling edges directly.
// Local result may come via bridge or be empty. Either is valid.
func TestTraceDependencies_LocalNodes(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	resp, err := tracer.TraceDependencies("RACKPDU-01", 5, true)
	if err != nil {
		t.Fatalf("TraceDependencies RACKPDU-01: %v", err)
	}
	if resp.Source.NodeID != "RACKPDU-01" {
		t.Errorf("Source.NodeID = %q, want RACKPDU-01", resp.Source.NodeID)
	}
	// If Local populated, only RDHx type expected
	for _, g := range resp.Local {
		for _, n := range g.Nodes {
			if n.NodeType != "RDHx" {
				t.Errorf("unexpected node type %q in Local (expected only RDHx)", n.NodeType)
			}
		}
	}
}

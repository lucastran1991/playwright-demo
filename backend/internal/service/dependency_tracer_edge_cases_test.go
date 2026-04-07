package service

import "testing"

// TestTraceFull_LevelsLimitUpstream verifies maxLevels restricts upstream depth.
func TestTraceFull_LevelsLimitUpstream(t *testing.T) {
	skipIfNoDB(t)
	// maxLevels=1: only RPP-01 at L1
	resp1, err := testTracer.TraceFull("RACKPDU-01", 1)
	if err != nil {
		t.Fatalf("maxLevels=1: %v", err)
	}
	if !hasNodeInUpstream(resp1, "RPP-01") {
		t.Error("expected RPP-01 at maxLevels=1")
	}
	if hasNodeInUpstream(resp1, "UPS-01") {
		t.Error("UPS-01 should NOT appear at maxLevels=1")
	}

	// maxLevels=2: both RPP-01 and UPS-01
	resp2, err := testTracer.TraceFull("RACKPDU-01", 2)
	if err != nil {
		t.Fatalf("maxLevels=2: %v", err)
	}
	if !hasNodeInUpstream(resp2, "RPP-01") || !hasNodeInUpstream(resp2, "UPS-01") {
		t.Error("expected both RPP-01 and UPS-01 at maxLevels=2")
	}
	if l := getUpstreamLevel(resp2, "RPP-01"); l != 1 {
		t.Errorf("RPP-01 level = %d, want 1", l)
	}
	if l := getUpstreamLevel(resp2, "UPS-01"); l != 2 {
		t.Errorf("UPS-01 level = %d, want 2", l)
	}
}

// TestTraceFull_LoadOnlyForCapacityNodes verifies Load is populated only for capacity nodes.
func TestTraceFull_LoadOnlyForCapacityNodes(t *testing.T) {
	skipIfNoDB(t)
	rppResp, err := testTracer.TraceFull("RPP-01", 5)
	if err != nil {
		t.Fatalf("RPP-01: %v", err)
	}
	if len(rppResp.Load) == 0 {
		t.Error("RPP-01 (capacity=true) should have non-empty Load")
	}

	pduResp, err := testTracer.TraceFull("RACKPDU-01", 5)
	if err != nil {
		t.Fatalf("RACKPDU-01: %v", err)
	}
	if len(pduResp.Load) != 0 {
		t.Errorf("RACKPDU-01 (capacity=false) should have empty Load, got %d", len(pduResp.Load))
	}
}

// TestTraceDependencies_BridgeFallback verifies CC-01 bridge behavior.
// No dep rules for Capacity Cell — empty upstream is correct.
func TestTraceDependencies_BridgeFallback(t *testing.T) {
	skipIfNoDB(t)
	resp, err := testTracer.TraceDependencies("CC-01", 5, false)
	if err != nil {
		t.Fatalf("CC-01: %v", err)
	}
	if resp.Source.NodeID != "CC-01" {
		t.Errorf("source: want CC-01, got %q", resp.Source.NodeID)
	}
}

// TestTraceDependencies_LocalNodes verifies RACKPDU-01 local dep handling.
func TestTraceDependencies_LocalNodes(t *testing.T) {
	skipIfNoDB(t)
	resp, err := testTracer.TraceDependencies("RACKPDU-01", 5, true)
	if err != nil {
		t.Fatalf("RACKPDU-01: %v", err)
	}
	if resp.Source.NodeID != "RACKPDU-01" {
		t.Errorf("source: want RACKPDU-01, got %q", resp.Source.NodeID)
	}
	// If Local populated, only RDHx type expected
	for _, g := range resp.Local {
		for _, n := range g.Nodes {
			if n.NodeType != "RDHx" {
				t.Errorf("unexpected type %q in Local (want RDHx)", n.NodeType)
			}
		}
	}
}

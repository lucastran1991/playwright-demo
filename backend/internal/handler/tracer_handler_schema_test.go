package handler

import "testing"

// TestTraceFull_LevelsCapped verifies ?levels=99 does not panic (capped to 10).
func TestTraceFull_LevelsCapped(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RACKPDU-01?levels=99")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil || len(resp.Data.Upstream) == 0 {
		t.Error("expected upstream nodes even with capped levels")
	}
}

// TestTraceFull_ResponseSchema verifies JSON field names match frontend expectations.
func TestTraceFull_ResponseSchema(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RACKPDU-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected data field")
	}

	src := resp.Data.Source
	if src.NodeID == "" {
		t.Error("source.node_id must not be empty")
	}
	if src.Name == "" {
		t.Error("source.name must not be empty")
	}
	if src.NodeType == "" {
		t.Error("source.node_type must not be empty")
	}
	if src.Topology == "" {
		t.Error("source.topology must not be empty")
	}

	for _, g := range resp.Data.Upstream {
		if g.Level <= 0 {
			t.Errorf("upstream group level must be >0, got %d", g.Level)
		}
		if g.Topology == "" {
			t.Error("upstream group topology must not be empty")
		}
		for _, n := range g.Nodes {
			if n.NodeID == "" || n.Name == "" || n.NodeType == "" || n.ID == 0 {
				t.Errorf("traced node missing fields: %+v", n)
			}
		}
	}
}

// TestTraceFull_EmptyUpstream verifies GEN-01 (top of chain) has empty upstream.
func TestTraceFull_EmptyUpstream(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/GEN-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data.Source.NodeID != "GEN-01" {
		t.Errorf("source: want GEN-01, got %q", resp.Data.Source.NodeID)
	}
	if len(resp.Data.Upstream) != 0 {
		t.Errorf("GEN-01 should have no upstream, got %d groups", len(resp.Data.Upstream))
	}
}

// TestTraceFull_LoadSection_CapacityNode verifies RPP-01 has load with Rack.
func TestTraceFull_LoadSection_CapacityNode(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RPP-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(resp.Data.Load) == 0 {
		t.Error("RPP-01 (capacity=true) should have non-empty load")
	}
	if findNodeInLocalGroups(resp.Data.Load, "RACK-01") == nil {
		t.Error("expected RACK-01 in load for RPP-01")
	}
}

// TestTraceFull_LoadSection_NonCapacityNode verifies RACKPDU-01 has empty load.
func TestTraceFull_LoadSection_NonCapacityNode(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RACKPDU-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(resp.Data.Load) != 0 {
		t.Errorf("RACKPDU-01 (capacity=false) should have empty load, got %d groups", len(resp.Data.Load))
	}
}

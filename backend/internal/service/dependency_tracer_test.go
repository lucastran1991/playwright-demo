package service

import (
	"strings"
	"testing"

	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/testutil"
)

// setupTracer creates a DependencyTracer backed by a real test DB.
func setupTracer(t *testing.T) (*DependencyTracer, func()) {
	t.Helper()
	db, cleanup := testutil.SetupTestDB(t)
	if err := testutil.TruncateAll(db); err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
	testutil.SeedTraceFixtures(t, db)
	repo := repository.NewTracerRepository(db)
	tracer := NewDependencyTracer(repo)
	return tracer, cleanup
}

// hasNodeInUpstream checks whether any upstream group contains the given nodeID.
func hasNodeInUpstream(resp *TraceResponse, nodeID string) bool {
	for _, g := range resp.Upstream {
		for _, n := range g.Nodes {
			if n.NodeID == nodeID {
				return true
			}
		}
	}
	return false
}

// hasNodeInDownstream checks whether any downstream group contains the given nodeID.
func hasNodeInDownstream(resp *TraceResponse, nodeID string) bool {
	for _, g := range resp.Downstream {
		for _, n := range g.Nodes {
			if n.NodeID == nodeID {
				return true
			}
		}
	}
	return false
}

// getUpstreamLevel returns the level at which nodeID appears in upstream, or -1 if absent.
func getUpstreamLevel(resp *TraceResponse, nodeID string) int {
	for _, g := range resp.Upstream {
		for _, n := range g.Nodes {
			if n.NodeID == nodeID {
				return g.Level
			}
		}
	}
	return -1
}

// TestTraceFull_MergesUpstreamAndDownstream verifies RPP-01 has both upstream and downstream.
func TestTraceFull_MergesUpstreamAndDownstream(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	resp, err := tracer.TraceFull("RPP-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RPP-01: %v", err)
	}
	if resp.Source.NodeID != "RPP-01" {
		t.Errorf("source NodeID = %q, want RPP-01", resp.Source.NodeID)
	}
	if !hasNodeInUpstream(resp, "UPS-01") {
		t.Error("expected UPS-01 in upstream, not found")
	}
	if !hasNodeInDownstream(resp, "RACKPDU-01") {
		t.Error("expected RACKPDU-01 in downstream, not found")
	}
}

// TestTraceFull_NodeNotFound verifies that a missing nodeID returns an error.
func TestTraceFull_NodeNotFound(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	_, err := tracer.TraceFull("DOES-NOT-EXIST", 5)
	if err == nil {
		t.Fatal("expected error for missing node, got nil")
	}
	if !strings.Contains(err.Error(), "node not found") {
		t.Errorf("error %q does not contain 'node not found'", err.Error())
	}
}

// TestTraceFull_NodeWithNoDependencyRules verifies GEN-01 returns empty upstream.
func TestTraceFull_NodeWithNoDependencyRules(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	resp, err := tracer.TraceFull("GEN-01", 5)
	if err != nil {
		t.Fatalf("TraceFull GEN-01: %v", err)
	}
	if len(resp.Upstream) != 0 {
		t.Errorf("expected empty upstream for GEN-01, got %d groups", len(resp.Upstream))
	}
}

// TestTraceFull_NodeWithNoImpactRules verifies RACKPDU-01 returns empty downstream.
func TestTraceFull_NodeWithNoImpactRules(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	resp, err := tracer.TraceFull("RACKPDU-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RACKPDU-01: %v", err)
	}
	if len(resp.Downstream) != 0 {
		t.Errorf("expected empty downstream for RACKPDU-01, got %d groups", len(resp.Downstream))
	}
}

// TestTraceFull_SourceNodeFields verifies all source fields are correct for RPP-01.
func TestTraceFull_SourceNodeFields(t *testing.T) {
	tracer, cleanup := setupTracer(t)
	defer cleanup()

	resp, err := tracer.TraceFull("RPP-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RPP-01: %v", err)
	}

	src := resp.Source
	if src.NodeID != "RPP-01" {
		t.Errorf("Source.NodeID = %q, want RPP-01", src.NodeID)
	}
	if src.Name != "RPP Panel 1" {
		t.Errorf("Source.Name = %q, want 'RPP Panel 1'", src.Name)
	}
	if src.NodeType != "RPP" {
		t.Errorf("Source.NodeType = %q, want RPP", src.NodeType)
	}
	if src.Topology != "Electrical System" {
		t.Errorf("Source.Topology = %q, want 'Electrical System'", src.Topology)
	}
}

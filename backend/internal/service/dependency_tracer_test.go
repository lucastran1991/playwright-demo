package service

import (
	"os"
	"strings"
	"testing"

	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/testutil"
)

// Package-level tracer seeded once in TestMain — all tests are read-only.
var testTracer *DependencyTracer

func TestMain(m *testing.M) {
	// Setup test DB if available; non-DB tests (helpers, CSV parsers) still run either way.
	db, cleanup := testutil.SetupTestDBForMain()
	if db != nil {
		testutil.TruncateAndSeedForMain(db)
		repo := repository.NewTracerRepository(db)
		testTracer = NewDependencyTracer(repo)
	}

	code := m.Run()
	if db != nil {
		cleanup()
	}
	os.Exit(code)
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

// getUpstreamLevel returns the level at which nodeID appears in upstream, or -1.
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

func skipIfNoDB(t *testing.T) {
	t.Helper()
	if testTracer == nil {
		t.Skip("PostgreSQL not available")
	}
}

// TestTraceFull_MergesUpstreamAndDownstream verifies RPP-01 has both.
func TestTraceFull_MergesUpstreamAndDownstream(t *testing.T) {
	skipIfNoDB(t)
	resp, err := testTracer.TraceFull("RPP-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RPP-01: %v", err)
	}
	if resp.Source.NodeID != "RPP-01" {
		t.Errorf("source: want RPP-01, got %q", resp.Source.NodeID)
	}
	if !hasNodeInUpstream(resp, "UPS-01") {
		t.Error("expected UPS-01 in upstream")
	}
	if !hasNodeInDownstream(resp, "RACKPDU-01") {
		t.Error("expected RACKPDU-01 in downstream")
	}
}

// TestTraceFull_NodeNotFound verifies error for missing node.
func TestTraceFull_NodeNotFound(t *testing.T) {
	skipIfNoDB(t)
	_, err := testTracer.TraceFull("DOES-NOT-EXIST", 5)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "node not found") {
		t.Errorf("error %q missing 'node not found'", err.Error())
	}
}

// TestTraceFull_NodeWithNoDependencyRules verifies GEN-01 has empty upstream.
func TestTraceFull_NodeWithNoDependencyRules(t *testing.T) {
	skipIfNoDB(t)
	resp, err := testTracer.TraceFull("GEN-01", 5)
	if err != nil {
		t.Fatalf("TraceFull GEN-01: %v", err)
	}
	if len(resp.Upstream) != 0 {
		t.Errorf("expected empty upstream, got %d groups", len(resp.Upstream))
	}
}

// TestTraceFull_NodeWithNoImpactRules verifies RACKPDU-01 has empty downstream.
func TestTraceFull_NodeWithNoImpactRules(t *testing.T) {
	skipIfNoDB(t)
	resp, err := testTracer.TraceFull("RACKPDU-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RACKPDU-01: %v", err)
	}
	if len(resp.Downstream) != 0 {
		t.Errorf("expected empty downstream, got %d groups", len(resp.Downstream))
	}
}

// TestTraceFull_SourceNodeFields verifies all source fields for RPP-01.
func TestTraceFull_SourceNodeFields(t *testing.T) {
	skipIfNoDB(t)
	resp, err := testTracer.TraceFull("RPP-01", 5)
	if err != nil {
		t.Fatalf("TraceFull RPP-01: %v", err)
	}
	src := resp.Source
	if src.NodeID != "RPP-01" {
		t.Errorf("NodeID = %q, want RPP-01", src.NodeID)
	}
	if src.Name != "RPP Panel 1" {
		t.Errorf("Name = %q, want 'RPP Panel 1'", src.Name)
	}
	if src.NodeType != "RPP" {
		t.Errorf("NodeType = %q, want RPP", src.NodeType)
	}
	if src.Topology != "Electrical System" {
		t.Errorf("Topology = %q, want 'Electrical System'", src.Topology)
	}
}

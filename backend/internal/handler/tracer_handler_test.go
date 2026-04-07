package handler

import (
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/service"
	"github.com/user/app/internal/testutil"
)

// Package-level router seeded once in TestMain — all tests are read-only.
var testRouter *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	db, cleanup := testutil.SetupTestDBForMain()
	if db == nil {
		os.Exit(0) // skip all — DB unavailable
	}
	defer cleanup()

	testutil.TruncateAndSeedForMain(db)

	repo := repository.NewTracerRepository(db)
	tracer := service.NewDependencyTracer(repo)
	h := NewTracerHandler(nil, tracer, repo, "")

	r := gin.New()
	trace := r.Group("/api/trace")
	trace.GET("/full/:nodeId", h.TraceFull)
	trace.GET("/dependencies/:nodeId", h.TraceDependencies)
	trace.GET("/impacts/:nodeId", h.TraceImpacts)
	testRouter = r

	os.Exit(m.Run())
}

// TestTraceFull_HappyPath_RPP verifies RPP-01 returns upstream + downstream.
func TestTraceFull_HappyPath_RPP(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RPP-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected data, got nil")
	}
	if resp.Data.Source.NodeID != "RPP-01" {
		t.Errorf("source node_id: want RPP-01, got %q", resp.Data.Source.NodeID)
	}
	if len(resp.Data.Upstream) == 0 {
		t.Error("expected non-empty upstream")
	}
	if len(resp.Data.Downstream) == 0 {
		t.Error("expected non-empty downstream")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "UPS-01") == nil {
		t.Error("expected UPS-01 in upstream")
	}
	if findNodeInLevelGroups(resp.Data.Downstream, "RACKPDU-01") == nil {
		t.Error("expected RACKPDU-01 in downstream")
	}
}

// TestTraceFull_HappyPath_RackPDU verifies upstream contains RPP at L1 and UPS at L2.
func TestTraceFull_HappyPath_RackPDU(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RACKPDU-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil || len(resp.Data.Upstream) == 0 {
		t.Fatal("expected non-empty upstream")
	}
	if g := levelGroupForNode(resp.Data.Upstream, "RPP-01"); g == nil {
		t.Error("expected RPP-01 in upstream")
	} else if g.Level != 1 {
		t.Errorf("RPP-01 level: want 1, got %d", g.Level)
	}
	if g := levelGroupForNode(resp.Data.Upstream, "UPS-01"); g == nil {
		t.Error("expected UPS-01 in upstream")
	} else if g.Level != 2 {
		t.Errorf("UPS-01 level: want 2, got %d", g.Level)
	}
}

// TestTraceFull_NotFound verifies 404 for unknown node.
func TestTraceFull_NotFound(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/NONEXISTENT-99")
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(resp.Error, "node not found") {
		t.Errorf("error: want 'node not found', got %q", resp.Error)
	}
}

// TestTraceFull_LevelsParam verifies ?levels=1 filters out L2 nodes.
func TestTraceFull_LevelsParam(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RACKPDU-01?levels=1")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "RPP-01") == nil {
		t.Error("expected RPP-01 at levels=1")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "UPS-01") != nil {
		t.Error("UPS-01 (L2) should not appear at levels=1")
	}
}

// TestTraceFull_DefaultLevels verifies default levels=2 returns both L1 and L2.
func TestTraceFull_DefaultLevels(t *testing.T) {
	w, resp := doTraceRequest(t, testRouter, "/api/trace/full/RACKPDU-01")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "RPP-01") == nil {
		t.Error("expected RPP-01 at default levels")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "UPS-01") == nil {
		t.Error("expected UPS-01 at default levels (L2)")
	}
}

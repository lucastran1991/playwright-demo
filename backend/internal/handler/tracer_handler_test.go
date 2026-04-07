package handler

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/service"
	"github.com/user/app/internal/testutil"
)

// setupTestRouter builds a minimal gin router with trace routes and real PostgreSQL.
// Returns the router and a cleanup function.
func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	t.Helper()
	db, cleanup := testutil.SetupTestDB(t)
	if err := testutil.TruncateAll(db); err != nil {
		cleanup()
		t.Fatalf("truncate: %v", err)
	}
	testutil.SeedTraceFixtures(t, db)

	repo := repository.NewTracerRepository(db)
	tracer := service.NewDependencyTracer(repo)
	h := NewTracerHandler(nil, tracer, repo, "") // nil ingestion svc — not testing that

	gin.SetMode(gin.TestMode)
	r := gin.New()
	trace := r.Group("/api/trace")
	trace.GET("/full/:nodeId", h.TraceFull)
	trace.GET("/dependencies/:nodeId", h.TraceDependencies)
	trace.GET("/impacts/:nodeId", h.TraceImpacts)

	return r, cleanup
}

// TestTraceFull_HappyPath_RPP verifies RPP-01 returns 200 with upstream (UPS-01) and downstream (RACKPDU-01).
func TestTraceFull_HappyPath_RPP(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	w, resp := doTraceRequest(t, router, "/api/trace/full/RPP-01")

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected data field, got nil")
	}
	if resp.Data.Source.NodeID != "RPP-01" {
		t.Errorf("source node_id: want RPP-01, got %q", resp.Data.Source.NodeID)
	}
	if resp.Data.Source.NodeType != "RPP" {
		t.Errorf("source node_type: want RPP, got %q", resp.Data.Source.NodeType)
	}
	if len(resp.Data.Upstream) == 0 {
		t.Error("expected non-empty upstream for RPP-01")
	}
	if len(resp.Data.Downstream) == 0 {
		t.Error("expected non-empty downstream for RPP-01")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "UPS-01") == nil {
		t.Error("expected UPS-01 in upstream for RPP-01")
	}
	if findNodeInLevelGroups(resp.Data.Downstream, "RACKPDU-01") == nil {
		t.Error("expected RACKPDU-01 in downstream for RPP-01")
	}
}

// TestTraceFull_HappyPath_RackPDU verifies RACKPDU-01 upstream contains RPP at L1 and UPS at L2.
func TestTraceFull_HappyPath_RackPDU(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	w, resp := doTraceRequest(t, router, "/api/trace/full/RACKPDU-01")

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected data field, got nil")
	}
	if len(resp.Data.Upstream) == 0 {
		t.Fatal("expected non-empty upstream for RACKPDU-01")
	}

	rppGroup := levelGroupForNode(resp.Data.Upstream, "RPP-01")
	if rppGroup == nil {
		t.Error("expected RPP-01 in upstream for RACKPDU-01")
	} else if rppGroup.Level != 1 {
		t.Errorf("RPP-01 upstream level: want 1, got %d", rppGroup.Level)
	}

	upsGroup := levelGroupForNode(resp.Data.Upstream, "UPS-01")
	if upsGroup == nil {
		t.Error("expected UPS-01 in upstream for RACKPDU-01")
	} else if upsGroup.Level != 2 {
		t.Errorf("UPS-01 upstream level: want 2, got %d", upsGroup.Level)
	}
}

// TestTraceFull_NotFound verifies that an unknown node ID returns 404 with "node not found" error.
func TestTraceFull_NotFound(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	w, resp := doTraceRequest(t, router, "/api/trace/full/NONEXISTENT-99")

	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	if !strings.Contains(resp.Error, "node not found") {
		t.Errorf("error message: want contains 'node not found', got %q", resp.Error)
	}
}

// TestTraceFull_LevelsParam verifies ?levels=1 returns RPP at L1 but not UPS at L2.
func TestTraceFull_LevelsParam(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	w, resp := doTraceRequest(t, router, "/api/trace/full/RACKPDU-01?levels=1")

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected data field, got nil")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "RPP-01") == nil {
		t.Error("expected RPP-01 in upstream at levels=1")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "UPS-01") != nil {
		t.Error("UPS-01 (L2) should not appear when levels=1")
	}
}

// TestTraceFull_DefaultLevels verifies that omitting ?levels defaults to 2.
func TestTraceFull_DefaultLevels(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	w, resp := doTraceRequest(t, router, "/api/trace/full/RACKPDU-01")

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if resp.Data == nil {
		t.Fatal("expected data field, got nil")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "RPP-01") == nil {
		t.Error("expected RPP-01 in upstream at default levels")
	}
	if findNodeInLevelGroups(resp.Data.Upstream, "UPS-01") == nil {
		t.Error("expected UPS-01 in upstream at default levels (should reach L2)")
	}
}

# Phase 4: Add /trace/full API Endpoint

## Context

- [Plan overview](plan.md)
- Depends on: Phase 2 + Phase 3

## Overview

- **Priority:** Medium
- **Status:** Complete
- **Description:** Add a combined `/api/trace/full/:nodeId` endpoint that returns upstream + local + downstream + load in a single response. Frontend uses one call for DAG view.

## Key Insights

Simple composition: call both TraceDependencies and TraceImpacts, merge into one TraceResponse. Keep existing separate endpoints for backward compatibility.

## Related Code Files

### Modify
- `backend/internal/service/dependency_tracer.go` — add TraceFull method
- `backend/internal/handler/tracer_handler.go` — add TraceFull handler
- `backend/internal/router/router.go` — register new route

## Implementation Steps

### Step 1: Add TraceFull Service Method

**File:** `backend/internal/service/dependency_tracer.go`

```go
// TraceFull returns combined dependency + impact trace in a single response.
func (t *DependencyTracer) TraceFull(nodeID string, maxLevels int) (*TraceResponse, error) {
    depResp, depErr := t.TraceDependencies(nodeID, maxLevels, true)
    impResp, impErr := t.TraceImpacts(nodeID, maxLevels)

    if depErr != nil && impErr != nil {
        return nil, depErr // both failed
    }

    // Use dep response as base (has source, upstream, local)
    resp := depResp
    if resp == nil {
        resp = impResp
    }
    if resp == nil {
        return nil, fmt.Errorf("node not found: %s", nodeID)
    }

    // Merge impact fields
    if impResp != nil {
        resp.Downstream = impResp.Downstream
        resp.Load = impResp.Load
    }

    return resp, nil
}
```

### Step 2: Add Handler

**File:** `backend/internal/handler/tracer_handler.go`

```go
// TraceFull handles GET /api/trace/full/:nodeId.
func (h *TracerHandler) TraceFull(c *gin.Context) {
    nodeID := c.Param("nodeId")
    levels := parseIntParam(c, "levels", 2, 10)

    result, err := h.tracer.TraceFull(nodeID, levels)
    if err != nil {
        if strings.Contains(err.Error(), "node not found") {
            response.Error(c, http.StatusNotFound, err.Error())
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to trace node")
        return
    }
    response.Success(c, http.StatusOK, result)
}
```

### Step 3: Register Route

**File:** `backend/internal/router/router.go`

Add inside the `trace` group (after line 58):

```go
trace.GET("/full/:nodeId", tracerHandler.TraceFull)
```

### Step 4: Verify

```bash
curl -s "http://localhost:8889/api/trace/full/RACK-R1-Z1-R1-01?levels=6" | python3 -c "
import sys, json
d = json.load(sys.stdin)['data']
print('source:', d['source']['node_type'])
print('upstream:', sum(len(g['nodes']) for g in d.get('upstream', [])))
print('local:', sum(len(g['nodes']) for g in d.get('local', [])))
print('downstream:', sum(len(g['nodes']) for g in d.get('downstream', [])))
print('load:', sum(len(g['nodes']) for g in d.get('load', [])))
"
```

## Todo List

- [x] Add TraceFull method to dependency_tracer.go
- [x] Add TraceFull handler to tracer_handler.go
- [x] Register GET /api/trace/full/:nodeId route
- [x] Verify combined response for Rack
- [x] Verify combined response for UPS (both sides populated)

## Success Criteria

1. Single API call returns all 4 sections: upstream, local, downstream, load
2. Response matches merged output of existing two endpoints
3. Existing /dependencies and /impacts still work unchanged

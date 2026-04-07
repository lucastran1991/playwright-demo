package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// -- JSON response types for unmarshalling trace API responses --

type apiResponse struct {
	Data  *traceData `json:"data"`
	Error string     `json:"error"`
}

type traceData struct {
	Source     sourceNode   `json:"source"`
	Upstream   []levelGroup `json:"upstream"`
	Local      []localGroup `json:"local"`
	Downstream []levelGroup `json:"downstream"`
	Load       []localGroup `json:"load"`
}

type sourceNode struct {
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	NodeType string `json:"node_type"`
	Topology string `json:"topology"`
}

type levelGroup struct {
	Level    int          `json:"level"`
	Topology string       `json:"topology"`
	Nodes    []tracedNode `json:"nodes"`
}

type localGroup struct {
	Topology string       `json:"topology"`
	Nodes    []tracedNode `json:"nodes"`
}

type tracedNode struct {
	ID           uint    `json:"id"`
	NodeID       string  `json:"node_id"`
	Name         string  `json:"name"`
	NodeType     string  `json:"node_type"`
	Level        int     `json:"level"`
	ParentNodeID *string `json:"parent_node_id"`
}

// -- Test helper functions --

// doTraceRequest fires a GET to the test router and decodes the JSON body.
func doTraceRequest(t *testing.T, router *gin.Engine, path string) (*httptest.ResponseRecorder, apiResponse) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var resp apiResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return w, resp
}

// findNodeInLevelGroups returns the first tracedNode with the given node_id across all level groups.
func findNodeInLevelGroups(groups []levelGroup, nodeID string) *tracedNode {
	for _, g := range groups {
		for i := range g.Nodes {
			if g.Nodes[i].NodeID == nodeID {
				return &g.Nodes[i]
			}
		}
	}
	return nil
}

// findNodeInLocalGroups returns the first tracedNode with the given node_id across all local groups.
func findNodeInLocalGroups(groups []localGroup, nodeID string) *tracedNode {
	for _, g := range groups {
		for i := range g.Nodes {
			if g.Nodes[i].NodeID == nodeID {
				return &g.Nodes[i]
			}
		}
	}
	return nil
}

// levelGroupForNode returns the levelGroup that contains the given node_id.
func levelGroupForNode(groups []levelGroup, nodeID string) *levelGroup {
	for i := range groups {
		for _, n := range groups[i].Nodes {
			if n.NodeID == nodeID {
				return &groups[i]
			}
		}
	}
	return nil
}

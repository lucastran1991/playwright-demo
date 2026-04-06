package handler

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/service"
	"github.com/user/app/pkg/response"
)

// TracerHandler handles model ingestion and dependency/impact tracing HTTP requests.
type TracerHandler struct {
	ingestionService *service.ModelIngestionService
	tracer           *service.DependencyTracer
	repo             *repository.TracerRepository
	modelDir         string
}

// NewTracerHandler creates a new TracerHandler instance.
func NewTracerHandler(
	svc *service.ModelIngestionService,
	tracer *service.DependencyTracer,
	repo *repository.TracerRepository,
	modelDir string,
) *TracerHandler {
	return &TracerHandler{ingestionService: svc, tracer: tracer, repo: repo, modelDir: modelDir}
}

// IngestModels handles POST /api/models/ingest -- triggers ingestion of 3 model CSVs.
func (h *TracerHandler) IngestModels(c *gin.Context) {
	summary, err := h.ingestionService.IngestAll(h.modelDir)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Model ingestion failed: "+err.Error())
		return
	}
	// Refresh cached topology lookups so trace API uses freshly ingested data
	h.tracer.RefreshLookups()
	response.Success(c, http.StatusOK, summary)
}

// ListCapacityNodes handles GET /api/models/capacity-nodes.
func (h *TracerHandler) ListCapacityNodes(c *gin.Context) {
	types, err := h.repo.ListCapacityNodeTypes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list capacity node types")
		return
	}
	response.Success(c, http.StatusOK, types)
}

// TraceDependencies handles GET /api/trace/dependencies/:nodeId.
func (h *TracerHandler) TraceDependencies(c *gin.Context) {
	nodeID := c.Param("nodeId")
	levels := parseIntParam(c, "levels", 2, 10)
	includeLocal := c.DefaultQuery("include_local", "true")

	result, err := h.tracer.TraceDependencies(nodeID, levels, strings.EqualFold(includeLocal, "true"))
	if err != nil {
		if strings.Contains(err.Error(), "node not found") {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to trace dependencies")
		return
	}
	response.Success(c, http.StatusOK, result)
}

// TraceImpacts handles GET /api/trace/impacts/:nodeId.
func (h *TracerHandler) TraceImpacts(c *gin.Context) {
	nodeID := c.Param("nodeId")
	levels := parseIntParam(c, "levels", 2, 10)

	result, err := h.tracer.TraceImpacts(nodeID, levels)
	if err != nil {
		if strings.Contains(err.Error(), "node not found") {
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to trace impacts")
		return
	}
	response.Success(c, http.StatusOK, result)
}

// TraceFull handles GET /api/trace/full/:nodeId — combined dependency + impact.
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

// TraceExportCSV handles GET /api/trace/full/:nodeId/export — CSV export of DAG tree.
func (h *TracerHandler) TraceExportCSV(c *gin.Context) {
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

	// Build node_id → name lookup from all traced nodes
	nameMap := make(map[string]string)
	src := result.Source
	nameMap[src.NodeID] = src.Name
	for _, groups := range [][]service.TraceLevelGroup{result.Upstream, result.Downstream} {
		for _, g := range groups {
			for _, n := range g.Nodes {
				nameMap[n.NodeID] = n.Name
			}
		}
	}
	for _, groups := range [][]service.TraceLocalGroup{result.Local, result.Load} {
		for _, g := range groups {
			for _, n := range g.Nodes {
				nameMap[n.NodeID] = n.Name
			}
		}
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="trace-%s.csv"`, nodeID))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"src_node_id", "src_node_name", "direction", "level", "topology", "node_id", "node_name", "node_type", "parent_node_id", "parent_node_name"})

	writeNodes := func(direction string, level int, topology string, nodes []repository.TracedNode) {
		for _, n := range nodes {
			parentID := ""
			parentName := ""
			if n.ParentNodeID != nil {
				parentID = *n.ParentNodeID
				parentName = nameMap[parentID]
			}
			w.Write([]string{src.NodeID, src.Name, direction, strconv.Itoa(level), topology, n.NodeID, n.Name, n.NodeType, parentID, parentName})
		}
	}

	// Source row
	w.Write([]string{src.NodeID, src.Name, "source", "0", src.Topology, src.NodeID, src.Name, src.NodeType, "", ""})

	for _, g := range result.Upstream {
		writeNodes("upstream", g.Level, g.Topology, g.Nodes)
	}
	for _, g := range result.Local {
		writeNodes("local", 0, g.Topology, g.Nodes)
	}
	for _, g := range result.Downstream {
		writeNodes("downstream", g.Level, g.Topology, g.Nodes)
	}
	for _, g := range result.Load {
		writeNodes("load", 0, g.Topology, g.Nodes)
	}

	w.Flush()
}

// TraceExportXLSX handles GET /api/trace/export/xlsx — bulk XLSX with one sheet per node type.
// Each sheet contains ALL instances of that type, traced at the given depth.
func (h *TracerHandler) TraceExportXLSX(c *gin.Context) {
	levels := parseIntParam(c, "levels", 2, 10)

	// Get all capacity node types
	capTypes, err := h.repo.ListCapacityNodeTypes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list node types")
		return
	}

	f := excelize.NewFile()
	defer f.Close()
	// Remove default "Sheet1"
	f.DeleteSheet("Sheet1")

	headers := []string{"src_node_id", "src_node_name", "direction", "level", "topology",
		"node_id", "node_name", "node_type", "parent_node_id", "parent_node_name"}

	for _, ct := range capTypes {
		if !ct.IsCapacityNode {
			continue
		}

		// Fetch all instances of this node type
		var nodes []struct {
			NodeID string `json:"node_id"`
			Name   string `json:"name"`
		}
		h.repo.DB().Raw(`SELECT node_id, name FROM blueprint_nodes WHERE node_type = ? ORDER BY node_id`, ct.NodeType).Scan(&nodes)

		if len(nodes) == 0 {
			continue
		}

		// Sheet name: max 31 chars
		sheetName := ct.NodeType
		if len(sheetName) > 31 {
			sheetName = sheetName[:31]
		}
		f.NewSheet(sheetName)

		// Write headers
		for col, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, 1)
			f.SetCellValue(sheetName, cell, h)
		}

		row := 2
		for _, node := range nodes {
			traceResp, err := h.tracer.TraceFull(node.NodeID, levels)
			if err != nil {
				log.Printf("WARNING: trace failed for %s: %v", node.NodeID, err)
				continue
			}

			// Build name lookup
			nameMap := make(map[string]string)
			nameMap[traceResp.Source.NodeID] = traceResp.Source.Name
			for _, groups := range [][]service.TraceLevelGroup{traceResp.Upstream, traceResp.Downstream} {
				for _, g := range groups {
					for _, n := range g.Nodes {
						nameMap[n.NodeID] = n.Name
					}
				}
			}
			for _, groups := range [][]service.TraceLocalGroup{traceResp.Local, traceResp.Load} {
				for _, g := range groups {
					for _, n := range g.Nodes {
						nameMap[n.NodeID] = n.Name
					}
				}
			}

			src := traceResp.Source

			writeRow := func(direction string, level int, topology string, n repository.TracedNode) {
				parentID, parentName := "", ""
				if n.ParentNodeID != nil {
					parentID = *n.ParentNodeID
					parentName = nameMap[parentID]
				}
				vals := []string{src.NodeID, src.Name, direction, strconv.Itoa(level), topology,
					n.NodeID, n.Name, n.NodeType, parentID, parentName}
				for col, v := range vals {
					cell, _ := excelize.CoordinatesToCellName(col+1, row)
					f.SetCellValue(sheetName, cell, v)
				}
				row++
			}

			// Source row
			srcNode := repository.TracedNode{NodeID: src.NodeID, Name: src.Name, NodeType: src.NodeType}
			writeRow("source", 0, src.Topology, srcNode)

			for _, g := range traceResp.Upstream {
				for _, n := range g.Nodes {
					writeRow("upstream", g.Level, g.Topology, n)
				}
			}
			for _, g := range traceResp.Local {
				for _, n := range g.Nodes {
					writeRow("local", 0, g.Topology, n)
				}
			}
			for _, g := range traceResp.Downstream {
				for _, n := range g.Nodes {
					writeRow("downstream", g.Level, g.Topology, n)
				}
			}
			for _, g := range traceResp.Load {
				for _, n := range g.Nodes {
					writeRow("load", 0, g.Topology, n)
				}
			}
		}
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", `attachment; filename="trace-all-models.xlsx"`)
	f.Write(c.Writer)
}

func parseIntParam(c *gin.Context, key string, defaultVal, maxVal int) int {
	if v, err := strconv.Atoi(c.Query(key)); err == nil && v > 0 {
		if v > maxVal {
			return maxVal
		}
		return v
	}
	return defaultVal
}

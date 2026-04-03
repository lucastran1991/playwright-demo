package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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

func parseIntParam(c *gin.Context, key string, defaultVal, maxVal int) int {
	if v, err := strconv.Atoi(c.Query(key)); err == nil && v > 0 {
		if v > maxVal {
			return maxVal
		}
		return v
	}
	return defaultVal
}

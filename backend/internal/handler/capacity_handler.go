package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/service"
	"github.com/user/app/pkg/response"
)

// CapacityHandler handles capacity-related HTTP requests.
type CapacityHandler struct {
	ingestionService *service.CapacityIngestionService
	repo             *repository.CapacityRepository
	csvPath          string
}

// NewCapacityHandler creates a new CapacityHandler.
func NewCapacityHandler(
	svc *service.CapacityIngestionService,
	repo *repository.CapacityRepository,
	csvPath string,
) *CapacityHandler {
	return &CapacityHandler{ingestionService: svc, repo: repo, csvPath: csvPath}
}

// IngestCapacity handles POST /api/capacity/ingest — triggers CSV ingestion + computation.
func (h *CapacityHandler) IngestCapacity(c *gin.Context) {
	summary, err := h.ingestionService.IngestCSV(h.csvPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Capacity ingestion failed: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, summary)
}

// GetNodeCapacity handles GET /api/capacity/nodes/:nodeId — single node metrics.
func (h *CapacityHandler) GetNodeCapacity(c *gin.Context) {
	nodeID := c.Param("nodeId")
	cap, err := h.repo.GetNodeCapacity(nodeID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get capacity data")
		return
	}
	if cap == nil {
		response.Error(c, http.StatusNotFound, "No capacity data for node: "+nodeID)
		return
	}
	response.Success(c, http.StatusOK, cap)
}

// GetSummary handles GET /api/capacity/summary — aggregate stats.
func (h *CapacityHandler) GetSummary(c *gin.Context) {
	summary, err := h.repo.GetCapacitySummary()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get capacity summary")
		return
	}
	response.Success(c, http.StatusOK, summary)
}

// ListCapacityNodes handles GET /api/capacity/nodes — paginated list with filters.
func (h *CapacityHandler) ListCapacityNodes(c *gin.Context) {
	nodeType := c.Query("type")
	minUtil := parseFloatParam(c, "min_utilization", 0)
	limit := parseIntParam(c, "limit", 50, 500)
	offset := parseIntParam(c, "offset", 0, 10000)

	nodes, total, err := h.repo.ListCapacityNodes(nodeType, minUtil, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list capacity nodes")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"nodes": nodes, "total": total})
}

func parseFloatParam(c *gin.Context, key string, defaultVal float64) float64 {
	if v, err := strconv.ParseFloat(c.Query(key), 64); err == nil && v >= 0 {
		return v
	}
	return defaultVal
}

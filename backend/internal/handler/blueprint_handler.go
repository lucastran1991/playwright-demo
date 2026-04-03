package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/app/internal/repository"
	"github.com/user/app/internal/service"
	"github.com/user/app/pkg/response"
	"gorm.io/gorm"
)

// BlueprintHandler handles blueprint HTTP requests.
type BlueprintHandler struct {
	ingestionService *service.BlueprintIngestionService
	repo             *repository.BlueprintRepository
	blueprintDir     string
}

// NewBlueprintHandler creates a new BlueprintHandler instance.
func NewBlueprintHandler(svc *service.BlueprintIngestionService, repo *repository.BlueprintRepository, dir string) *BlueprintHandler {
	return &BlueprintHandler{ingestionService: svc, repo: repo, blueprintDir: dir}
}

// Ingest handles POST /api/blueprints/ingest -- triggers full CSV ingestion.
func (h *BlueprintHandler) Ingest(c *gin.Context) {
	summary, err := h.ingestionService.IngestAll(h.blueprintDir)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Ingestion failed: "+err.Error())
		return
	}
	response.Success(c, http.StatusOK, summary)
}

// ListTypes handles GET /api/blueprints/types -- returns all blueprint domains.
func (h *BlueprintHandler) ListTypes(c *gin.Context) {
	types, err := h.repo.ListTypes()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list types")
		return
	}
	response.Success(c, http.StatusOK, types)
}

// ListNodes handles GET /api/blueprints/nodes -- returns nodes with optional type filter.
func (h *BlueprintHandler) ListNodes(c *gin.Context) {
	typeSlug := c.Query("type")
	limit, offset := parsePagination(c)

	nodes, total, err := h.repo.ListNodes(typeSlug, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list nodes")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": nodes, "total": total, "limit": limit, "offset": offset})
}

// GetNode handles GET /api/blueprints/nodes/:nodeId -- returns a node with memberships.
func (h *BlueprintHandler) GetNode(c *gin.Context) {
	nodeID := c.Param("nodeId")
	node, memberships, err := h.repo.GetNodeByNodeID(nodeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "Node not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Failed to get node")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": node, "memberships": memberships})
}

// ListEdges handles GET /api/blueprints/edges -- returns edges for a blueprint type.
func (h *BlueprintHandler) ListEdges(c *gin.Context) {
	typeSlug := c.Query("type")
	if typeSlug == "" {
		response.Error(c, http.StatusBadRequest, "Query param 'type' is required")
		return
	}
	limit, offset := parsePagination(c)

	edges, total, err := h.repo.ListEdges(typeSlug, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to list edges")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": edges, "total": total, "limit": limit, "offset": offset})
}

// GetTree handles GET /api/blueprints/tree/:typeSlug -- returns recursive tree.
func (h *BlueprintHandler) GetTree(c *gin.Context) {
	typeSlug := c.Param("typeSlug")
	tree, err := h.repo.GetTree(typeSlug)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to get tree")
		return
	}
	response.Success(c, http.StatusOK, tree)
}

// parsePagination extracts limit/offset from query params with defaults.
func parsePagination(c *gin.Context) (int, int) {
	limit := 100
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}
	return limit, offset
}

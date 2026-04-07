package repository

import (
	"github.com/user/app/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CapacityRepository handles node_variables DB operations.
type CapacityRepository struct {
	db *gorm.DB
}

// NewCapacityRepository creates a new CapacityRepository.
func NewCapacityRepository(db *gorm.DB) *CapacityRepository {
	return &CapacityRepository{db: db}
}

// UpsertNodeVariable upserts a single node variable using ON CONFLICT.
func (r *CapacityRepository) UpsertNodeVariable(tx *gorm.DB, nv *model.NodeVariable) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "node_id"}, {Name: "variable_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "unit", "source", "updated_at"}),
	}).Create(nv).Error
}

// BulkUpsert upserts a batch of node variables.
func (r *CapacityRepository) BulkUpsert(tx *gorm.DB, vars []model.NodeVariable) error {
	if len(vars) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "node_id"}, {Name: "variable_name"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "unit", "source", "updated_at"}),
	}).CreateInBatches(vars, 100).Error
}

// GetNodeVariables returns all variables for a given node_id.
func (r *CapacityRepository) GetNodeVariables(nodeID string) ([]model.NodeVariable, error) {
	var vars []model.NodeVariable
	err := r.db.Where("node_id = ?", nodeID).Find(&vars).Error
	return vars, err
}

// DeleteBySource deletes all node_variables with given source (for re-computation).
func (r *CapacityRepository) DeleteBySource(tx *gorm.DB, source string) error {
	return tx.Where("source = ?", source).Delete(&model.NodeVariable{}).Error
}

// GetVariableMap returns map[nodeID]map[varName]float64 for a given source.
func (r *CapacityRepository) GetVariableMap(source string) (map[string]map[string]float64, error) {
	var vars []model.NodeVariable
	q := r.db.Model(&model.NodeVariable{})
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if err := q.Find(&vars).Error; err != nil {
		return nil, err
	}
	result := make(map[string]map[string]float64)
	for _, v := range vars {
		if result[v.NodeID] == nil {
			result[v.NodeID] = make(map[string]float64)
		}
		result[v.NodeID][v.VariableName] = v.Value
	}
	return result, nil
}

// NodeIDWithDBID pairs a node_id string with its DB primary key.
type NodeIDWithDBID struct {
	ID     uint   `json:"id"`
	NodeID string `json:"node_id"`
}

// GetNodeIDsByType returns all blueprint_node IDs + DB IDs for a given node_type.
func (r *CapacityRepository) GetNodeIDsByType(nodeType string) ([]NodeIDWithDBID, error) {
	var nodes []NodeIDWithDBID
	err := r.db.Raw(`SELECT id, node_id FROM blueprint_nodes WHERE node_type = ?`, nodeType).Scan(&nodes).Error
	return nodes, err
}

// NodeCapacity is the API-facing capacity summary for a single node.
type NodeCapacity struct {
	NodeID   string             `json:"node_id"`
	NodeType string             `json:"node_type"`
	Name     string             `json:"name"`
	Capacity map[string]float64 `json:"capacity"`
	Units    map[string]string  `json:"units"`
}

// GetNodeCapacity returns capacity metrics for a single node.
func (r *CapacityRepository) GetNodeCapacity(nodeID string) (*NodeCapacity, error) {
	var vars []model.NodeVariable
	if err := r.db.Where("node_id = ?", nodeID).Find(&vars).Error; err != nil {
		return nil, err
	}
	if len(vars) == 0 {
		return nil, nil
	}

	// Look up node metadata
	var node struct {
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		NodeType string `json:"node_type"`
	}
	if err := r.db.Raw(`SELECT node_id, name, node_type FROM blueprint_nodes WHERE node_id = ? LIMIT 1`, nodeID).Scan(&node).Error; err != nil {
		return nil, err
	}

	cap := &NodeCapacity{
		NodeID:   nodeID,
		NodeType: node.NodeType,
		Name:     node.Name,
		Capacity: make(map[string]float64),
		Units:    make(map[string]string),
	}
	for _, v := range vars {
		cap.Capacity[v.VariableName] = v.Value
		cap.Units[v.VariableName] = v.Unit
	}
	return cap, nil
}

// CapacitySummary holds aggregate stats across all capacity nodes.
type CapacitySummary struct {
	TotalNodes      int     `json:"total_nodes"`
	AvgUtilization  float64 `json:"avg_utilization_pct"`
	OverloadedNodes int     `json:"overloaded_nodes"`
	HighUtilNodes   int     `json:"high_util_nodes"`
	TotalCapacity   float64 `json:"total_capacity_kw"`
	TotalLoad       float64 `json:"total_load_kw"`
}

// GetCapacitySummary returns aggregate stats across all nodes with capacity data.
func (r *CapacityRepository) GetCapacitySummary() (*CapacitySummary, error) {
	var summary CapacitySummary
	err := r.db.Raw(`
		SELECT
			COUNT(DISTINCT node_id) as total_nodes,
			COALESCE(AVG(CASE WHEN variable_name = 'utilization_pct' THEN value END), 0) as avg_utilization,
			COUNT(DISTINCT CASE WHEN variable_name = 'utilization_pct' AND value > 100 THEN node_id END) as overloaded_nodes,
			COUNT(DISTINCT CASE WHEN variable_name = 'utilization_pct' AND value > 80 AND value <= 100 THEN node_id END) as high_util_nodes,
			COALESCE(SUM(CASE WHEN variable_name = 'rated_capacity' AND source = 'csv_import' THEN value ELSE 0 END), 0) as total_capacity,
			COALESCE(SUM(CASE WHEN variable_name = 'allocated_load' AND source = 'csv_import' THEN value ELSE 0 END), 0) as total_load
		FROM node_variables
	`).Scan(&summary).Error
	return &summary, err
}

// ListCapacityNodes returns paginated nodes with capacity data, optionally filtered.
func (r *CapacityRepository) ListCapacityNodes(nodeType string, minUtil float64, limit, offset int) ([]NodeCapacity, int64, error) {
	var nodeIDs []string
	var total int64

	if nodeType != "" {
		r.db.Raw(`SELECT COUNT(DISTINCT nv.node_id) FROM node_variables nv
			JOIN blueprint_nodes bn ON bn.node_id = nv.node_id
			WHERE bn.node_type = ?`, nodeType).Scan(&total)
		r.db.Raw(`SELECT DISTINCT nv.node_id FROM node_variables nv
			JOIN blueprint_nodes bn ON bn.node_id = nv.node_id
			WHERE bn.node_type = ?
			ORDER BY nv.node_id LIMIT ? OFFSET ?`, nodeType, limit, offset).Scan(&nodeIDs)
	} else if minUtil > 0 {
		r.db.Raw(`SELECT COUNT(DISTINCT node_id) FROM node_variables
			WHERE variable_name = 'utilization_pct' AND value >= ?`, minUtil).Scan(&total)
		r.db.Raw(`SELECT DISTINCT node_id FROM node_variables
			WHERE variable_name = 'utilization_pct' AND value >= ?
			ORDER BY node_id LIMIT ? OFFSET ?`, minUtil, limit, offset).Scan(&nodeIDs)
	} else {
		r.db.Model(&model.NodeVariable{}).Select("COUNT(DISTINCT node_id)").Scan(&total)
		r.db.Model(&model.NodeVariable{}).Select("DISTINCT node_id").
			Order("node_id").Limit(limit).Offset(offset).Scan(&nodeIDs)
	}

	if len(nodeIDs) == 0 {
		return nil, total, nil
	}

	// Batch-load all variables for matched nodes (avoids N+1)
	capMap, err := r.GetCapacityMapForNodes(nodeIDs)
	if err != nil {
		return nil, 0, err
	}

	// Batch-load node metadata
	var nodes []struct {
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		NodeType string `json:"node_type"`
	}
	r.db.Raw(`SELECT node_id, name, node_type FROM blueprint_nodes WHERE node_id IN ?`, nodeIDs).Scan(&nodes)
	metaMap := make(map[string]struct{ Name, NodeType string }, len(nodes))
	for _, n := range nodes {
		metaMap[n.NodeID] = struct{ Name, NodeType string }{n.Name, n.NodeType}
	}

	// Assemble results
	results := make([]NodeCapacity, 0, len(nodeIDs))
	for _, nid := range nodeIDs {
		vars := capMap[nid]
		if vars == nil {
			continue
		}
		meta := metaMap[nid]
		nc := NodeCapacity{
			NodeID:   nid,
			NodeType: meta.NodeType,
			Name:     meta.Name,
			Capacity: vars,
			Units:    make(map[string]string),
		}
		results = append(results, nc)
	}
	return results, total, nil
}

// GetCapacityMapForNodes returns capacity data for a batch of node IDs.
func (r *CapacityRepository) GetCapacityMapForNodes(nodeIDs []string) (map[string]map[string]float64, error) {
	if len(nodeIDs) == 0 {
		return nil, nil
	}
	var vars []model.NodeVariable
	if err := r.db.Where("node_id IN ?", nodeIDs).Find(&vars).Error; err != nil {
		return nil, err
	}
	result := make(map[string]map[string]float64)
	for _, v := range vars {
		if result[v.NodeID] == nil {
			result[v.NodeID] = make(map[string]float64)
		}
		result[v.NodeID][v.VariableName] = v.Value
	}
	return result, nil
}

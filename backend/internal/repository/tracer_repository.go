package repository

import (
	"github.com/user/app/internal/model"
	"gorm.io/gorm"
)

// TracedNode represents a node found during dependency/impact tracing.
type TracedNode struct {
	ID       uint   `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	NodeType string `json:"node_type"`
	Level    int    `json:"level"`
}

// TracerRepository handles queries for dependency/impact tracing.
type TracerRepository struct {
	db *gorm.DB
}

// NewTracerRepository creates a new TracerRepository instance.
func NewTracerRepository(db *gorm.DB) *TracerRepository {
	return &TracerRepository{db: db}
}

// FindUpstreamNodes walks parent edges recursively in a given topology up to maxLevel hops.
func (r *TracerRepository) FindUpstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error) {
	var nodes []TracedNode
	err := r.db.Raw(`
		WITH RECURSIVE upstream AS (
			SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
			FROM blueprint_edges be
			JOIN blueprint_nodes bn ON bn.id = be.from_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE be.to_node_id = ? AND bt.slug = ?

			UNION ALL

			SELECT bn.id, bn.node_id, bn.name, bn.node_type, u.level + 1
			FROM upstream u
			JOIN blueprint_edges be ON be.to_node_id = u.id
			JOIN blueprint_nodes bn ON bn.id = be.from_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE bt.slug = ? AND u.level < ?
		)
		SELECT DISTINCT id, node_id, name, node_type, MIN(level) as level
		FROM upstream
		GROUP BY id, node_id, name, node_type
		ORDER BY level, node_type, node_id
	`, sourceDBID, typeSlug, typeSlug, maxLevel).Scan(&nodes).Error
	return nodes, err
}

// FindDownstreamNodes walks child edges recursively in a given topology up to maxLevel hops.
func (r *TracerRepository) FindDownstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error) {
	var nodes []TracedNode
	err := r.db.Raw(`
		WITH RECURSIVE downstream AS (
			SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
			FROM blueprint_edges be
			JOIN blueprint_nodes bn ON bn.id = be.to_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE be.from_node_id = ? AND bt.slug = ?

			UNION ALL

			SELECT bn.id, bn.node_id, bn.name, bn.node_type, d.level + 1
			FROM downstream d
			JOIN blueprint_edges be ON be.from_node_id = d.id
			JOIN blueprint_nodes bn ON bn.id = be.to_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE bt.slug = ? AND d.level < ?
		)
		SELECT DISTINCT id, node_id, name, node_type, MIN(level) as level
		FROM downstream
		GROUP BY id, node_id, name, node_type
		ORDER BY level, node_type, node_id
	`, sourceDBID, typeSlug, typeSlug, maxLevel).Scan(&nodes).Error
	return nodes, err
}

// FindLocalNodes returns direct edge neighbors of a node in a given topology.
func (r *TracerRepository) FindLocalNodes(sourceDBID uint, typeSlug string) ([]TracedNode, error) {
	var nodes []TracedNode
	err := r.db.Raw(`
		SELECT DISTINCT bn.id, bn.node_id, bn.name, bn.node_type, 0 as level
		FROM blueprint_edges be
		JOIN blueprint_nodes bn ON (bn.id = be.from_node_id OR bn.id = be.to_node_id)
		JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
		WHERE bt.slug = ?
		  AND (be.from_node_id = ? OR be.to_node_id = ?)
		  AND bn.id != ?
		ORDER BY bn.node_type, bn.node_id
	`, typeSlug, sourceDBID, sourceDBID, sourceDBID).Scan(&nodes).Error
	return nodes, err
}

// GetDependencyRules returns all dependency rules for a given node type.
func (r *TracerRepository) GetDependencyRules(nodeType string) ([]model.DependencyRule, error) {
	var rules []model.DependencyRule
	err := r.db.Where("node_type = ?", nodeType).Order("topological_relationship, upstream_level").Find(&rules).Error
	return rules, err
}

// GetImpactRules returns all impact rules for a given node type.
func (r *TracerRepository) GetImpactRules(nodeType string) ([]model.ImpactRule, error) {
	var rules []model.ImpactRule
	err := r.db.Where("node_type = ?", nodeType).Order("topological_relationship, downstream_level").Find(&rules).Error
	return rules, err
}

// ListCapacityNodeTypes returns all capacity node type metadata.
func (r *TracerRepository) ListCapacityNodeTypes() ([]model.CapacityNodeType, error) {
	var types []model.CapacityNodeType
	err := r.db.Order("topology, node_type").Find(&types).Error
	return types, err
}

// FindNodeByStringID looks up a blueprint node by its string node_id.
func (r *TracerRepository) FindNodeByStringID(nodeID string) (*model.BlueprintNode, error) {
	var node model.BlueprintNode
	if err := r.db.Where("node_id = ?", nodeID).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// ListBlueprintTypes returns all blueprint types for topology-to-slug mapping.
func (r *TracerRepository) ListBlueprintTypes() ([]model.BlueprintType, error) {
	var types []model.BlueprintType
	if err := r.db.Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

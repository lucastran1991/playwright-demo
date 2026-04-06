package repository

import (
	"github.com/user/app/internal/model"
	"gorm.io/gorm"
)

// TracedNode represents a node found during dependency/impact tracing.
type TracedNode struct {
	ID           uint    `json:"id"`
	NodeID       string  `json:"node_id"`
	Name         string  `json:"name"`
	NodeType     string  `json:"node_type"`
	Level        int     `json:"level"`
	ParentNodeID *string `json:"parent_node_id,omitempty"`
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
			SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level,
			       child.node_id as parent_node_id
			FROM blueprint_edges be
			JOIN blueprint_nodes bn ON bn.id = be.from_node_id
			JOIN blueprint_nodes child ON child.id = be.to_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE be.to_node_id = ? AND bt.slug = ?

			UNION ALL

			SELECT bn.id, bn.node_id, bn.name, bn.node_type, u.level + 1,
			       u.node_id as parent_node_id
			FROM upstream u
			JOIN blueprint_edges be ON be.to_node_id = u.id
			JOIN blueprint_nodes bn ON bn.id = be.from_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE bt.slug = ? AND u.level < ?
		)
		SELECT id, node_id, name, node_type, MIN(level) as level, parent_node_id
		FROM upstream
		GROUP BY id, node_id, name, node_type, parent_node_id
		ORDER BY level, node_type, node_id
	`, sourceDBID, typeSlug, typeSlug, maxLevel).Scan(&nodes).Error
	return nodes, err
}

// FindDownstreamNodes walks child edges recursively in a given topology up to maxLevel hops.
func (r *TracerRepository) FindDownstreamNodes(sourceDBID uint, typeSlug string, maxLevel int) ([]TracedNode, error) {
	var nodes []TracedNode
	err := r.db.Raw(`
		WITH RECURSIVE downstream AS (
			SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level,
			       parent.node_id as parent_node_id
			FROM blueprint_edges be
			JOIN blueprint_nodes bn ON bn.id = be.to_node_id
			JOIN blueprint_nodes parent ON parent.id = be.from_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE be.from_node_id = ? AND bt.slug = ?

			UNION ALL

			SELECT bn.id, bn.node_id, bn.name, bn.node_type, d.level + 1,
			       d.node_id as parent_node_id
			FROM downstream d
			JOIN blueprint_edges be ON be.from_node_id = d.id
			JOIN blueprint_nodes bn ON bn.id = be.to_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE bt.slug = ? AND d.level < ?
		)
		SELECT id, node_id, name, node_type, MIN(level) as level, parent_node_id
		FROM downstream
		GROUP BY id, node_id, name, node_type, parent_node_id
		ORDER BY level, node_type, node_id
	`, sourceDBID, typeSlug, typeSlug, maxLevel).Scan(&nodes).Error
	return nodes, err
}

// FindSpatialAncestorsOfType walks up spatial edges from the given node IDs
// and returns distinct ancestor nodes whose node_type is in the given set.
// Used to find Load nodes (Rack, Row, Zone) that are spatial parents of
// electrical downstream nodes (RACKPDU, etc.).
func (r *TracerRepository) FindSpatialAncestorsOfType(nodeDBIDs []uint, nodeTypes []string) ([]TracedNode, error) {
	if len(nodeDBIDs) == 0 || len(nodeTypes) == 0 {
		return nil, nil
	}
	var nodes []TracedNode
	err := r.db.Raw(`
		WITH RECURSIVE ancestors AS (
			SELECT bn.id, bn.node_id, bn.name, bn.node_type, 1 as level
			FROM blueprint_edges be
			JOIN blueprint_nodes bn ON bn.id = be.from_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE be.to_node_id IN ?
			  AND bt.slug = 'spatial-topology'

			UNION ALL

			SELECT bn.id, bn.node_id, bn.name, bn.node_type, a.level + 1
			FROM ancestors a
			JOIN blueprint_edges be ON be.to_node_id = a.id
			JOIN blueprint_nodes bn ON bn.id = be.from_node_id
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE bt.slug = 'spatial-topology' AND a.level < 5
		)
		SELECT DISTINCT ON (id) id, node_id, name, node_type, MIN(level) as level
		FROM ancestors
		WHERE node_type IN ?
		GROUP BY id, node_id, name, node_type
		ORDER BY id, level
	`, nodeDBIDs, nodeTypes).Scan(&nodes).Error
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

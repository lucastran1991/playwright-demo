package repository

import (
	"github.com/user/app/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlueprintRepository handles database operations for blueprint models.
type BlueprintRepository struct {
	db *gorm.DB
}

// NewBlueprintRepository creates a new BlueprintRepository instance.
func NewBlueprintRepository(db *gorm.DB) *BlueprintRepository {
	return &BlueprintRepository{db: db}
}

// UpsertType inserts or updates a BlueprintType by name.
func (r *BlueprintRepository) UpsertType(tx *gorm.DB, bt *model.BlueprintType) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "folder_name", "updated_at"}),
	}).Create(bt).Error
}

// UpsertNode inserts or updates a BlueprintNode by node_id.
// Ensures node.ID is populated after upsert for downstream edge resolution.
func (r *BlueprintRepository) UpsertNode(tx *gorm.DB, node *model.BlueprintNode) error {
	err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "node_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "node_type", "node_role", "updated_at"}),
	}).Create(node).Error
	if err != nil {
		return err
	}
	if node.ID == 0 {
		return tx.Where("node_id = ?", node.NodeID).First(node).Error
	}
	return nil
}

// UpsertMembership inserts or updates a BlueprintNodeMembership.
func (r *BlueprintRepository) UpsertMembership(tx *gorm.DB, m *model.BlueprintNodeMembership) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "blueprint_type_id"}, {Name: "blueprint_node_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"org_path", "updated_at"}),
	}).Create(m).Error
}

// UpsertEdge inserts or ignores a BlueprintEdge (no update needed on conflict).
func (r *BlueprintRepository) UpsertEdge(tx *gorm.DB, e *model.BlueprintEdge) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "blueprint_type_id"}, {Name: "from_node_id"}, {Name: "to_node_id"}},
		DoNothing: true,
	}).Create(e).Error
}

// FindNodeByNodeID looks up a BlueprintNode by its string node_id.
func (r *BlueprintRepository) FindNodeByNodeID(tx *gorm.DB, nodeID string) (*model.BlueprintNode, error) {
	var node model.BlueprintNode
	if err := tx.Where("node_id = ?", nodeID).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// ListTypes returns all blueprint types.
func (r *BlueprintRepository) ListTypes() ([]model.BlueprintType, error) {
	var types []model.BlueprintType
	if err := r.db.Order("name").Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

// ListNodes returns nodes with optional type filter and pagination.
func (r *BlueprintRepository) ListNodes(typeSlug string, limit, offset int) ([]model.BlueprintNode, int64, error) {
	query := r.db.Model(&model.BlueprintNode{})
	if typeSlug != "" {
		query = query.Where("id IN (?)",
			r.db.Model(&model.BlueprintNodeMembership{}).Select("blueprint_node_id").
				Joins("JOIN blueprint_types ON blueprint_types.id = blueprint_node_memberships.blueprint_type_id").
				Where("blueprint_types.slug = ?", typeSlug),
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var nodes []model.BlueprintNode
	if err := query.Limit(limit).Offset(offset).Order("node_id").Find(&nodes).Error; err != nil {
		return nil, 0, err
	}
	return nodes, total, nil
}

// GetNodeByNodeID returns a single node with its memberships preloaded.
func (r *BlueprintRepository) GetNodeByNodeID(nodeID string) (*model.BlueprintNode, []model.BlueprintNodeMembership, error) {
	var node model.BlueprintNode
	if err := r.db.Where("node_id = ?", nodeID).First(&node).Error; err != nil {
		return nil, nil, err
	}

	var memberships []model.BlueprintNodeMembership
	if err := r.db.Where("blueprint_node_id = ?", node.ID).
		Preload("BlueprintType").Find(&memberships).Error; err != nil {
		return nil, nil, err
	}
	return &node, memberships, nil
}

// ListEdges returns edges for a blueprint type with pagination.
func (r *BlueprintRepository) ListEdges(typeSlug string, limit, offset int) ([]model.BlueprintEdge, int64, error) {
	query := r.db.Model(&model.BlueprintEdge{}).
		Joins("JOIN blueprint_types ON blueprint_types.id = blueprint_edges.blueprint_type_id").
		Where("blueprint_types.slug = ?", typeSlug)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var edges []model.BlueprintEdge
	if err := r.db.Preload("FromNode").Preload("ToNode").
		Joins("JOIN blueprint_types ON blueprint_types.id = blueprint_edges.blueprint_type_id").
		Where("blueprint_types.slug = ?", typeSlug).
		Limit(limit).Offset(offset).Find(&edges).Error; err != nil {
		return nil, 0, err
	}
	return edges, total, nil
}

// TreeNode represents a node in the recursive tree response.
type TreeNode struct {
	NodeID   string     `json:"node_id"`
	Name     string     `json:"name"`
	NodeType string     `json:"node_type"`
	OrgPath  string     `json:"org_path"`
	Children []TreeNode `json:"children,omitempty"`
}

// GetTree returns the full tree for a blueprint type using recursive CTE.
func (r *BlueprintRepository) GetTree(typeSlug string) ([]TreeNode, error) {
	// Find root nodes (nodes that are parents but not children in this domain)
	type flatRow struct {
		NodeID       string
		Name         string
		NodeType     string
		OrgPath      string
		ParentNodeID *string
	}

	var rows []flatRow
	err := r.db.Raw(`
		WITH domain_edges AS (
			SELECT be.from_node_id, be.to_node_id
			FROM blueprint_edges be
			JOIN blueprint_types bt ON bt.id = be.blueprint_type_id
			WHERE bt.slug = ?
		)
		SELECT
			bn.node_id,
			bn.name,
			bn.node_type,
			bnm.org_path,
			parent.node_id AS parent_node_id
		FROM blueprint_node_memberships bnm
		JOIN blueprint_types bt ON bt.id = bnm.blueprint_type_id
		JOIN blueprint_nodes bn ON bn.id = bnm.blueprint_node_id
		LEFT JOIN domain_edges de ON de.to_node_id = bn.id
		LEFT JOIN blueprint_nodes parent ON parent.id = de.from_node_id
		WHERE bt.slug = ?
		ORDER BY bnm.org_path
	`, typeSlug, typeSlug).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// Build tree from flat rows
	nodeMap := make(map[string]*TreeNode)
	childrenMap := make(map[string][]string) // parent_node_id -> child node_ids

	for _, row := range rows {
		nodeMap[row.NodeID] = &TreeNode{
			NodeID:   row.NodeID,
			Name:     row.Name,
			NodeType: row.NodeType,
			OrgPath:  row.OrgPath,
		}
		if row.ParentNodeID != nil {
			childrenMap[*row.ParentNodeID] = append(childrenMap[*row.ParentNodeID], row.NodeID)
		}
	}

	// Recursive build with cycle detection
	visited := make(map[string]bool)
	var buildChildren func(nodeID string) []TreeNode
	buildChildren = func(nodeID string) []TreeNode {
		if visited[nodeID] {
			return nil
		}
		visited[nodeID] = true
		childIDs := childrenMap[nodeID]
		if len(childIDs) == 0 {
			return nil
		}
		children := make([]TreeNode, 0, len(childIDs))
		for _, cid := range childIDs {
			if n, ok := nodeMap[cid]; ok {
				child := *n
				child.Children = buildChildren(cid)
				children = append(children, child)
			}
		}
		return children
	}

	// Find roots (nodes with no parent in this domain)
	var roots []TreeNode
	for _, row := range rows {
		if row.ParentNodeID == nil {
			if n, ok := nodeMap[row.NodeID]; ok {
				root := *n
				root.Children = buildChildren(row.NodeID)
				roots = append(roots, root)
			}
		}
	}
	return roots, nil
}


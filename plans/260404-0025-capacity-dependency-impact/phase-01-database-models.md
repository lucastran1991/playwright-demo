# Phase 1: Database Models + Migration

## Context Links
- Brainstorm: `plans/reports/brainstorm-260404-0025-capacity-dependency-impact-model.md`
- Existing model pattern: `backend/internal/model/blueprint_type.go`
- Migration: `backend/internal/database/database.go`

## Overview
- **Priority**: P1 (blocks all other phases)
- **Status**: completed
- **Description**: Create 3 GORM model structs and register them in AutoMigrate

## Key Insights
- These are **type-level metadata tables**, not instance topology -- separate from blueprint_* tables
- Tiny tables (24 + 147 + 118 rows) -- no performance concerns
- Composite unique constraints needed for idempotent upsert

## Requirements

### Functional
- `capacity_node_types` table: stores which node types are capacity domains/constraints
- `dependency_rules` table: type-level upstream/local dependency rules
- `impact_rules` table: type-level downstream/load impact rules
- All tables support upsert (ON CONFLICT DO UPDATE)

### Non-functional
- Follow existing model conventions (timestamps, primaryKey, json tags)
- Composite unique indexes for upsert conflict targets

## Architecture

### Table: capacity_node_types
| Column | Type | Constraint |
|--------|------|-----------|
| id | uint | PK |
| node_type | string(100) | unique, not null |
| topology | string(100) | not null |
| is_capacity_node | bool | not null, default false |
| active_constraint | bool | not null, default false |
| created_at | time | auto |
| updated_at | time | auto |

### Table: dependency_rules
| Column | Type | Constraint |
|--------|------|-----------|
| id | uint | PK |
| node_type | string(100) | composite unique, index, not null |
| dependency_node_type | string(100) | composite unique, not null |
| relationship_type | string(20) | not null (always "Dependency") |
| topological_relationship | string(20) | not null ("Upstream" or "Local") |
| upstream_level | *int | nullable (null for Local) |
| created_at | time | auto |
| updated_at | time | auto |

### Table: impact_rules
| Column | Type | Constraint |
|--------|------|-----------|
| id | uint | PK |
| node_type | string(100) | composite unique, index, not null |
| impact_node_type | string(100) | composite unique, not null |
| topological_relationship | string(20) | not null ("Downstream" or "Load") |
| downstream_level | *int | nullable (null for Load) |
| created_at | time | auto |
| updated_at | time | auto |

## Related Code Files

### Files to Create
- `backend/internal/model/capacity_node_type.go`
- `backend/internal/model/dependency_rule.go`
- `backend/internal/model/impact_rule.go`

### Files to Modify
- `backend/internal/database/database.go` -- add 3 models to `Migrate()` func

## Implementation Steps

### 1. Create `backend/internal/model/capacity_node_type.go`
```go
package model

import "time"

type CapacityNodeType struct {
    ID               uint      `gorm:"primaryKey" json:"id"`
    NodeType         string    `gorm:"uniqueIndex;size:100;not null" json:"node_type"`
    Topology         string    `gorm:"size:100;not null" json:"topology"`
    IsCapacityNode   bool      `gorm:"not null;default:false" json:"is_capacity_node"`
    ActiveConstraint bool      `gorm:"not null;default:false" json:"active_constraint"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}
```

### 2. Create `backend/internal/model/dependency_rule.go`
```go
package model

import "time"

type DependencyRule struct {
    ID                      uint      `gorm:"primaryKey" json:"id"`
    NodeType                string    `gorm:"uniqueIndex:idx_dep_rule_type_pair;index;size:100;not null" json:"node_type"`
    DependencyNodeType      string    `gorm:"uniqueIndex:idx_dep_rule_type_pair;size:100;not null" json:"dependency_node_type"`
    RelationshipType        string    `gorm:"size:20;not null" json:"relationship_type"`
    TopologicalRelationship string    `gorm:"size:20;not null" json:"topological_relationship"`
    UpstreamLevel           *int      `gorm:"" json:"upstream_level"`
    CreatedAt               time.Time `json:"created_at"`
    UpdatedAt               time.Time `json:"updated_at"`
}
```

### 3. Create `backend/internal/model/impact_rule.go`
```go
package model

import "time"

type ImpactRule struct {
    ID                      uint      `gorm:"primaryKey" json:"id"`
    NodeType                string    `gorm:"uniqueIndex:idx_impact_rule_type_pair;index;size:100;not null" json:"node_type"`
    ImpactNodeType          string    `gorm:"uniqueIndex:idx_impact_rule_type_pair;size:100;not null" json:"impact_node_type"`
    TopologicalRelationship string    `gorm:"size:20;not null" json:"topological_relationship"`
    DownstreamLevel         *int      `gorm:"" json:"downstream_level"`
    CreatedAt               time.Time `json:"created_at"`
    UpdatedAt               time.Time `json:"updated_at"`
}
```

### 4. Update `backend/internal/database/database.go`
Add to the `Migrate()` function:
```go
&model.CapacityNodeType{},
&model.DependencyRule{},
&model.ImpactRule{},
```

## Todo List
- [x] Create capacity_node_type.go
- [x] Create dependency_rule.go
- [x] Create impact_rule.go
- [x] Update database.go AutoMigrate
- [x] Verify `go build` compiles

## Success Criteria
- All 3 models compile without errors
- `AutoMigrate` creates tables with correct columns and indexes
- Composite unique indexes work for upsert conflict resolution

## Risk Assessment
- **Low**: straightforward GORM model definitions following existing patterns
- **Low**: AutoMigrate handles schema creation; no manual SQL needed

## Next Steps
- Phase 2 depends on these models for CSV parsing targets

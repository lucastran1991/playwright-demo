# Phase 2: Migration Updates

## Context Links
- [Plan Overview](plan.md)
- [Phase 1: Models](phase-01-database-models.md)
- Existing: `backend/internal/database/database.go`

## Overview
- **Priority**: P1 (blocks ingestion)
- **Status**: completed
- **Description**: Add 4 new models to GORM AutoMigrate call

## Key Insights
- Current `Migrate()` only handles `model.User`
- GORM AutoMigrate creates tables + indexes from struct tags
- Composite unique indexes need `gorm:"uniqueIndex:idx_name"` tags on multiple fields

## Requirements

### Functional
- Add BlueprintType, BlueprintNode, BlueprintNodeMembership, BlueprintEdge to AutoMigrate
- Tables created with correct indexes and constraints

### Non-functional
- database.go stays under 200 lines
- Backward compatible -- existing User migration unaffected

## Related Code Files

### Files to Modify
- `backend/internal/database/database.go` -- add 4 models to Migrate()

### Files to Reference
- Phase 1 model files

## Implementation Steps

1. Import `model` package (already imported)
2. Add 4 new model structs to `db.AutoMigrate()` call:
   ```go
   return db.AutoMigrate(
       &model.User{},
       &model.BlueprintType{},
       &model.BlueprintNode{},
       &model.BlueprintNodeMembership{},
       &model.BlueprintEdge{},
   )
   ```
3. Verify composite unique indexes are created by GORM from struct tags

## Todo List
- [x] Update Migrate() in database.go
- [x] Run app to verify tables created
- [x] Check indexes in PostgreSQL: `\d blueprint_nodes` etc.

## Success Criteria
- All 4 tables created in PostgreSQL on app startup
- Unique indexes present on: `blueprint_nodes.node_id`, `blueprint_types.name`, `blueprint_types.slug`
- Composite unique indexes on memberships and edges

## Risk Assessment
- **Low**: AutoMigrate is additive, won't break existing User table

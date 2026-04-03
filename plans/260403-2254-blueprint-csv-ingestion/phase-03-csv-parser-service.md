# Phase 3: CSV Parser Service

## Context Links
- [Plan Overview](plan.md)
- CSV location: `blueprint/Node & Edge/`
- 6 folders: Cooling system_Blueprint, Deployment Blueprint, Electrical system_Blueprint, Operational infrastructure_Blueprint, Spatial Topology_Blueprint, Whitespace Blueprint

## Overview
- **Priority**: P1 (blocks ingestion)
- **Status**: completed
- **Description**: Generic CSV parser that auto-discovers blueprint folders and parses Nodes.csv + Edges.csv into typed structs

## Key Insights
- All 6 domains have identical CSV schemas
- Folder names become blueprint type identifiers
- Files are named exactly `Nodes.csv` and `Edges.csv` in each subfolder
- CSV uses comma delimiter, has header row
- Data is small (~5K rows total) -- load entire file into memory is fine

## Requirements

### Functional
- Auto-discover subdirectories under a configurable base path
- Parse Nodes.csv: columns `Node ID, Node Name, Node Role, Org Path, Node Type`
- Parse Edges.csv: columns `From Node Name, From Node ID, From Node Org Path, To Node Name, To Node ID, To Node Org Path`
- Validate CSV headers before parsing
- Return typed structs, not raw string maps
- Skip empty rows gracefully

### Non-functional
- File under 200 lines
- No hardcoded folder names -- discover dynamically
- Return descriptive errors for missing/malformed files

## Architecture

```
blueprint_csv_parser.go
  ├── DiscoverDomains(basePath) -> []DomainFolder
  ├── ParseNodesCSV(filePath) -> []NodeRow
  └── ParseEdgesCSV(filePath) -> []EdgeRow

DomainFolder { Name, Path }
NodeRow { NodeID, Name, Role, OrgPath, NodeType }
EdgeRow { FromName, FromNodeID, FromOrgPath, ToName, ToNodeID, ToOrgPath }
```

## Related Code Files

### Files to Create
- `backend/internal/service/blueprint_csv_parser.go`

### Files to Reference
- `backend/internal/service/auth_service.go` -- service conventions

## Implementation Steps

1. Define raw row structs:
   ```go
   type NodeRow struct {
       NodeID   string
       Name     string
       Role     string
       OrgPath  string
       NodeType string
   }
   
   type EdgeRow struct {
       FromName    string
       FromNodeID  string
       FromOrgPath string
       ToName      string
       ToNodeID    string
       ToOrgPath   string
   }
   
   type DomainFolder struct {
       Name string  // folder name e.g. "Cooling system_Blueprint"
       Path string  // full path to folder
   }
   ```

2. Implement `DiscoverDomains(basePath string) ([]DomainFolder, error)`
   - `os.ReadDir(basePath)` 
   - Filter for directories only
   - Return sorted list

3. Implement `ParseNodesCSV(filePath string) ([]NodeRow, error)`
   - Open file, create `csv.NewReader`
   - Read and validate header row (must match expected 5 columns)
   - Parse remaining rows into NodeRow structs
   - Skip rows where NodeID is empty
   - Return slice

4. Implement `ParseEdgesCSV(filePath string) ([]EdgeRow, error)`
   - Same pattern as nodes
   - Validate header (6 columns)
   - Parse into EdgeRow structs
   - Skip rows where FromNodeID or ToNodeID is empty

## Todo List
- [x] Create blueprint_csv_parser.go
- [x] Implement DiscoverDomains
- [x] Implement ParseNodesCSV with header validation
- [x] Implement ParseEdgesCSV with header validation
- [x] Verify file compiles

## Success Criteria
- Discovers all 6 domain folders dynamically
- Parses all Nodes.csv files without errors
- Parses all Edges.csv files without errors
- Returns descriptive error if CSV headers don't match expected schema
- Handles empty rows gracefully

## Risk Assessment
- **Low**: CSV format might change -- header validation catches this early
- **Low**: File encoding issues -- Go's csv package handles standard CSV well
- **Mitigation**: Log which file/row caused parse errors

## Security Considerations
- Validate file paths to prevent directory traversal (only read from configured base path)
- No user-supplied file paths in this phase (base path from config)

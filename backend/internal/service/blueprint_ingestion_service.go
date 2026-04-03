package service

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
	"gorm.io/gorm"
)

// IngestionSummary holds the results of a full blueprint ingestion run.
type IngestionSummary struct {
	DomainsProcessed int      `json:"domains_processed"`
	NodesUpserted    int      `json:"nodes_upserted"`
	EdgesUpserted    int      `json:"edges_upserted"`
	EdgesSkipped     int      `json:"edges_skipped"`
	Errors           []string `json:"errors,omitempty"`
	DurationMs       int64    `json:"duration_ms"`
}

// domainResult holds counts from a single domain ingestion.
type domainResult struct {
	nodes   int
	edges   int
	skipped int
}

// BlueprintIngestionService orchestrates CSV parsing and database upsert for all domains.
type BlueprintIngestionService struct {
	repo *repository.BlueprintRepository
	db   *gorm.DB
}

// NewBlueprintIngestionService creates a new ingestion service.
func NewBlueprintIngestionService(repo *repository.BlueprintRepository, db *gorm.DB) *BlueprintIngestionService {
	return &BlueprintIngestionService{repo: repo, db: db}
}

// IngestAll discovers all blueprint domains and ingests their CSV data.
func (s *BlueprintIngestionService) IngestAll(basePath string) (*IngestionSummary, error) {
	start := time.Now()
	summary := &IngestionSummary{}

	domains, err := DiscoverDomains(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to discover domains: %w", err)
	}

	for _, domain := range domains {
		result, err := s.ingestDomain(domain)
		if err != nil {
			summary.Errors = append(summary.Errors, fmt.Sprintf("%s: %v", domain.Name, err))
			log.Printf("WARNING: failed to ingest domain %s: %v", domain.Name, err)
			continue
		}
		summary.DomainsProcessed++
		summary.NodesUpserted += result.nodes
		summary.EdgesUpserted += result.edges
		summary.EdgesSkipped += result.skipped
	}

	summary.DurationMs = time.Since(start).Milliseconds()
	return summary, nil
}

// ingestDomain processes a single blueprint domain within a transaction.
func (s *BlueprintIngestionService) ingestDomain(domain DomainFolder) (*domainResult, error) {
	result := &domainResult{}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		bt := &model.BlueprintType{
			Name:       FolderToName(domain.Name),
			Slug:       FolderToSlug(domain.Name),
			FolderName: domain.Name,
		}
		if err := s.repo.UpsertType(tx, bt); err != nil {
			return fmt.Errorf("upsert type: %w", err)
		}

		nodesFile, err := FindCSVFile(domain.Path, "node")
		if err != nil {
			return fmt.Errorf("find nodes CSV: %w", err)
		}
		nodeRows, err := ParseNodesCSV(nodesFile)
		if err != nil {
			return fmt.Errorf("parse nodes CSV %s: %w", filepath.Base(nodesFile), err)
		}

		// Build node ID -> DB ID map for edge resolution
		nodeIDMap := make(map[string]uint, len(nodeRows))
		for _, row := range nodeRows {
			node := &model.BlueprintNode{
				NodeID:   row.NodeID,
				Name:     row.Name,
				NodeType: row.NodeType,
				NodeRole: row.Role,
			}
			if err := s.repo.UpsertNode(tx, node); err != nil {
				return fmt.Errorf("upsert node %s: %w", row.NodeID, err)
			}
			nodeIDMap[row.NodeID] = node.ID

			membership := &model.BlueprintNodeMembership{
				BlueprintTypeID: bt.ID,
				BlueprintNodeID: node.ID,
				OrgPath:         row.OrgPath,
			}
			if err := s.repo.UpsertMembership(tx, membership); err != nil {
				return fmt.Errorf("upsert membership for %s: %w", row.NodeID, err)
			}
		}
		result.nodes = len(nodeRows)

		edgesFile, err := FindCSVFile(domain.Path, "edge")
		if err != nil {
			return fmt.Errorf("find edges CSV: %w", err)
		}
		edgeRows, err := ParseEdgesCSV(edgesFile)
		if err != nil {
			return fmt.Errorf("parse edges CSV %s: %w", filepath.Base(edgesFile), err)
		}

		for _, row := range edgeRows {
			fromID, ok := nodeIDMap[row.FromNodeID]
			if !ok {
				fromNode, err := s.repo.FindNodeByNodeID(tx, row.FromNodeID)
				if err != nil {
					log.Printf("WARNING: edge source node %s not found, skipping", row.FromNodeID)
					result.skipped++
					continue
				}
				fromID = fromNode.ID
			}

			toID, ok := nodeIDMap[row.ToNodeID]
			if !ok {
				toNode, err := s.repo.FindNodeByNodeID(tx, row.ToNodeID)
				if err != nil {
					log.Printf("WARNING: edge target node %s not found, skipping", row.ToNodeID)
					result.skipped++
					continue
				}
				toID = toNode.ID
			}

			edge := &model.BlueprintEdge{
				BlueprintTypeID: bt.ID,
				FromNodeID:      fromID,
				ToNodeID:        toID,
			}
			if err := s.repo.UpsertEdge(tx, edge); err != nil {
				return fmt.Errorf("upsert edge %s->%s: %w", row.FromNodeID, row.ToNodeID, err)
			}
			result.edges++
		}

		return nil
	})

	return result, err
}

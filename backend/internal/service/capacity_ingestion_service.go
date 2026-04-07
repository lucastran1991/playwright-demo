package service

import (
	"fmt"
	"time"

	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
	"gorm.io/gorm"
)

// CapacityIngestionSummary holds results of capacity CSV ingestion.
type CapacityIngestionSummary struct {
	NodesProcessed    int      `json:"nodes_processed"`
	VariablesUpserted int      `json:"variables_upserted"`
	ComputedVariables int      `json:"computed_variables"`
	RowsSkipped       int      `json:"rows_skipped"`
	Errors            []string `json:"errors,omitempty"`
	DurationMs        int64    `json:"duration_ms"`
}

// CapacityIngestionService orchestrates capacity CSV parsing and storage.
type CapacityIngestionService struct {
	repo       *repository.CapacityRepository
	calculator *LoadCapacityCalculator
	db         *gorm.DB
}

// NewCapacityIngestionService creates a new CapacityIngestionService.
func NewCapacityIngestionService(
	repo *repository.CapacityRepository,
	calculator *LoadCapacityCalculator,
	db *gorm.DB,
) *CapacityIngestionService {
	return &CapacityIngestionService{repo: repo, calculator: calculator, db: db}
}

// IngestCSV parses the capacity rack-load-flow CSV and upserts raw variables.
// After raw import, triggers bottom-up aggregation via calculator.
func (s *CapacityIngestionService) IngestCSV(filePath string) (*CapacityIngestionSummary, error) {
	start := time.Now()
	summary := &CapacityIngestionSummary{}

	rows, err := ParseCapacityFlowCSV(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse capacity CSV: %w", err)
	}

	// Upsert raw variables in transaction
	var totalVars int
	err = s.db.Transaction(func(tx *gorm.DB) error {
		for _, row := range rows {
			batch := make([]model.NodeVariable, 0, len(row.Variables))
			for _, v := range row.Variables {
				batch = append(batch, model.NodeVariable{
					NodeID:       row.NodeID,
					VariableName: v.VariableName,
					Value:        v.Value,
					Unit:         v.Unit,
					Source:       "csv_import",
				})
			}
			if err := s.repo.BulkUpsert(tx, batch); err != nil {
				return fmt.Errorf("upsert vars for %s: %w", row.NodeID, err)
			}
			totalVars += len(batch)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	summary.NodesProcessed = len(rows)
	summary.VariablesUpserted = totalVars

	// Compute bottom-up aggregates
	if s.calculator != nil {
		computeSummary, err := s.calculator.ComputeAll()
		if err != nil {
			summary.Errors = append(summary.Errors, "compute aggregates: "+err.Error())
		} else {
			summary.ComputedVariables = computeSummary.VariablesComputed
		}
	}

	summary.DurationMs = time.Since(start).Milliseconds()
	return summary, nil
}

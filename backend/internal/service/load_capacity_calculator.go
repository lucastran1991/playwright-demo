package service

import (
	"log"

	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
	"gorm.io/gorm"
)

// ComputeSummary holds results of bottom-up aggregation.
type ComputeSummary struct {
	VariablesComputed int `json:"variables_computed"`
	NodesProcessed    int `json:"nodes_processed"`
}

// LoadCapacityCalculator computes bottom-up load aggregation across the topology.
type LoadCapacityCalculator struct {
	capRepo    *repository.CapacityRepository
	tracerRepo *repository.TracerRepository
	db         *gorm.DB
}

// NewLoadCapacityCalculator creates a new calculator instance.
func NewLoadCapacityCalculator(
	capRepo *repository.CapacityRepository,
	tracerRepo *repository.TracerRepository,
	db *gorm.DB,
) *LoadCapacityCalculator {
	return &LoadCapacityCalculator{capRepo: capRepo, tracerRepo: tracerRepo, db: db}
}

// aggregationConfig defines how to aggregate load for a given node type.
type aggregationConfig struct {
	NodeType   string
	LoadVar    string // which rack variable to sum (allocated_load, allocated_air_load, etc.)
	CapVar     string // which variable holds this node's own capacity
	HasOwnCap  bool   // true if node has its own rated/design capacity from CSV
}

var aggregationConfigs = []aggregationConfig{
	// Power chain
	{"RPP", "allocated_load", "rated_capacity", true},
	{"Room PDU", "allocated_load", "rated_capacity", true},
	{"UPS", "allocated_load", "rated_capacity", true},
	// Cooling chain — air
	{"Air Zone", "allocated_air_load", "design_capacity", true},
	{"Air Cooling Unit", "allocated_air_load", "rated_capacity", true},
	// Cooling chain — liquid
	{"Liquid Loop", "allocated_liquid_load", "design_capacity", true},
	{"CDU", "allocated_liquid_load", "rated_capacity", true},
	{"RDHx", "allocated_liquid_load", "rated_capacity", true},
	{"DTC", "allocated_liquid_load", "rated_capacity", true},
	// Spatial/whitespace aggregates — use total IT load
	{"Capacity Cell", "allocated_load", "design_capacity", true},
	{"Room Bundle", "allocated_load", "design_capacity", true},
	{"UPS Bundle", "allocated_load", "design_capacity", true},
	{"Room PDU Bundle", "allocated_load", "design_capacity", true},
	// Row has no own capacity but tracks load
	{"Row", "allocated_load", "", false},
}

// ComputeAll deletes prior computed values, then computes and stores all aggregates.
func (c *LoadCapacityCalculator) ComputeAll() (*ComputeSummary, error) {
	summary := &ComputeSummary{}

	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Delete all previously computed values
		if err := c.capRepo.DeleteBySource(tx, "computed"); err != nil {
			return err
		}

		// Step 2: Load all csv_import variables into memory
		varMap, err := c.capRepo.GetVariableMap("csv_import")
		if err != nil {
			return err
		}

		var computed []model.NodeVariable

		// Step 3: Rack-level derived metrics
		rackNodes, err := c.capRepo.GetNodeIDsByType("Rack")
		if err != nil {
			return err
		}
		for _, rack := range rackNodes {
			nodeVars := varMap[rack.NodeID]
			if nodeVars == nil {
				continue
			}
			rated := nodeVars["rated_capacity"]
			allocated := nodeVars["allocated_load"]

			available := rated - allocated
			computed = append(computed, model.NodeVariable{
				NodeID: rack.NodeID, VariableName: "available_capacity",
				Value: available, Unit: "kW", Source: "computed",
			})

			utilPct := 0.0
			if rated > 0 {
				utilPct = (allocated / rated) * 100
			}
			computed = append(computed, model.NodeVariable{
				NodeID: rack.NodeID, VariableName: "utilization_pct",
				Value: utilPct, Unit: "%", Source: "computed",
			})
		}

		// Step 4: Aggregate capacity nodes — find descendant Racks via spatial topology
		for _, cfg := range aggregationConfigs {
			nodes, err := c.capRepo.GetNodeIDsByType(cfg.NodeType)
			if err != nil {
				log.Printf("load_calculator: GetNodeIDsByType(%s) error: %v", cfg.NodeType, err)
				continue
			}
			if len(nodes) == 0 {
				continue
			}

			for _, node := range nodes {
				// Find descendant Rack nodes via spatial hierarchy
				descendantRacks, err := c.tracerRepo.FindSpatialDescendantsOfType(
					[]uint{node.ID}, []string{"Rack"},
				)
				if err != nil {
					log.Printf("load_calculator: FindSpatialDescendantsOfType for %s error: %v", node.NodeID, err)
					continue
				}

				// Sum the target load variable from descendant racks
				var totalLoad float64
				for _, rack := range descendantRacks {
					rackVars := varMap[rack.NodeID]
					if rackVars == nil {
						continue
					}
					totalLoad += rackVars[cfg.LoadVar]
				}

				// Store computed allocated_load
				computed = append(computed, model.NodeVariable{
					NodeID: node.NodeID, VariableName: "allocated_load",
					Value: totalLoad, Unit: "kW", Source: "computed",
				})

				// Compute available + utilization if node has its own capacity
				if cfg.HasOwnCap && cfg.CapVar != "" {
					nodeVars := varMap[node.NodeID]
					capacity := 0.0
					if nodeVars != nil {
						capacity = nodeVars[cfg.CapVar]
					}

					available := capacity - totalLoad
					computed = append(computed, model.NodeVariable{
						NodeID: node.NodeID, VariableName: "available_capacity",
						Value: available, Unit: "kW", Source: "computed",
					})

					utilPct := 0.0
					if capacity > 0 {
						utilPct = (totalLoad / capacity) * 100
					}
					computed = append(computed, model.NodeVariable{
						NodeID: node.NodeID, VariableName: "utilization_pct",
						Value: utilPct, Unit: "%", Source: "computed",
					})
				}

				summary.NodesProcessed++
			}
		}

		// Step 5: Bulk upsert all computed variables
		if err := c.capRepo.BulkUpsert(tx, computed); err != nil {
			return err
		}
		summary.VariablesComputed = len(computed)
		return nil
	})

	return summary, err
}

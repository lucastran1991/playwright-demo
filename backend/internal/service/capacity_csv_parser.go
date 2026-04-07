package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// CapacityFlowRow represents a parsed row from the capacity rack-load-flow CSV.
type CapacityFlowRow struct {
	NodeID    string
	NodeType  string
	Name      string
	Variables []CapacityVariable
}

// CapacityVariable holds a single parsed variable with standardized name and unit.
type CapacityVariable struct {
	VariableName string
	Value        float64
	Unit         string
}

// columnMapping maps a CSV column header to a standardized variable name and unit.
type columnMapping struct {
	CSVHeader    string
	VariableName string
	Unit         string
}

// nodeTypeMappings defines which CSV columns map to which variables per node type.
var nodeTypeMappings = map[string][]columnMapping{
	"Rack": {
		{"Rack_Circuit_Capacity_(design)", "design_capacity", "kW"},
		{"Rack_Circuit_Capacity_(rated)", "rated_capacity", "kW"},
		{"Rack_LiquidCool_Fraction", "liquid_cool_fraction", "fraction"},
		{"Rack_AirCool_Fraction", "air_cool_fraction", "fraction"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
		{"Allocated_LiquidCool_Load", "allocated_liquid_load", "kW"},
		{"Allocated_AirCool_Load", "allocated_air_load", "kW"},
	},
	"RPP": {
		{"RPP_Panel_Capacity_(design)", "design_capacity", "kW"},
		{"RPP_Panel_Capacity_(rated)", "rated_capacity", "kW"},
		{"RPP_BreakerPole_Capacity", "breaker_pole_capacity", "count"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
		{"Rack_Count", "rack_count", "count"},
	},
	"Room PDU": {
		{"RPDU_Design_Capacity", "design_capacity", "kW"},
		{"RPDU_Operational_Cap", "rated_capacity", "kW"},
		{"RPDU_Transformer_Rating", "transformer_rating", "kVA"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
	},
	"UPS": {
		{"UPS_Design_Capacity", "design_capacity", "kW"},
		{"UPS_Rated_Capacity", "rated_capacity", "kW"},
		{"UPS_Operational_Cap", "operational_capacity", "kW"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
	},
	"Air Zone": {
		{"AirZone_Cooling_Capacity", "design_capacity", "kW"},
		{"Allocated_AirCool_Load", "allocated_air_load", "kW"},
		{"Rack_Count", "rack_count", "count"},
		{"Rack_DeltaT_(design)", "rack_delta_t", "degC"},
		{"Maximum_Air_Flow", "max_air_flow", "CFM"},
		{"ColdAisle_Supply_Temp", "cold_aisle_supply_temp", "degC"},
		{"Rack_Inlet_Setpoint", "rack_inlet_setpoint", "degC"},
	},
	"Liquid Loop": {
		{"LL_Cooling_Capacity", "design_capacity", "kW"},
		{"Allocated_LiquidCool_Load", "allocated_liquid_load", "kW"},
		{"Rack_Count", "rack_count", "count"},
		{"Loop_Supply_Setpoint", "loop_supply_setpoint", "degC"},
		{"Loop_DeltaT_(design)", "loop_delta_t", "degC"},
		{"Maximum_Flow_Rate", "max_flow_rate", "LPM"},
	},
	"Air Cooling Unit": {
		{"Rated_Cooling_Capacity", "rated_capacity", "kW"},
		{"Pump_Capacity", "pump_capacity", "kW"},
	},
	"CDU": {
		{"Rated_Cooling_Capacity", "rated_capacity", "kW"},
		{"Pump_Capacity", "pump_capacity", "kW"},
		{"Allocated_LiquidCool_Load", "allocated_liquid_load", "kW"},
	},
	"RDHx": {
		{"RDHx_MaxHeat_Removal", "rated_capacity", "kW"},
	},
	"DTC": {
		{"DTC_MaxCool_Capacity", "rated_capacity", "kW"},
	},
	"Row": {
		{"Allocated_ITLoad", "allocated_load", "kW"},
		{"Rack_Count", "rack_count", "count"},
	},
	"Capacity Cell": {
		{"Capacity_Envelope", "design_capacity", "kW"},
		{"Power_Capacity", "power_capacity", "kW"},
		{"Thermal_Capacity", "thermal_capacity", "kW"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
		{"Rack_Count", "rack_count", "count"},
	},
	"Room Bundle": {
		{"Capacity_Envelope", "design_capacity", "kW"},
		{"Power_Capacity", "power_capacity", "kW"},
		{"Thermal_Capacity", "thermal_capacity", "kW"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
		{"Rack_Count", "rack_count", "count"},
	},
	"Room PDU Bundle": {
		{"Capacity_Envelope", "design_capacity", "kW"},
		{"Power_Capacity", "power_capacity", "kW"},
		{"Thermal_Capacity", "thermal_capacity", "kW"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
	},
	"UPS Bundle": {
		{"Capacity_Envelope", "design_capacity", "kW"},
		{"Power_Capacity", "power_capacity", "kW"},
		{"Thermal_Capacity", "thermal_capacity", "kW"},
		{"Allocated_ITLoad", "allocated_load", "kW"},
	},
}

// ParseCapacityFlowCSV parses the ISET capacity rack-load-flow CSV file.
// It maps CSV columns to standardized variable names per node type.
func ParseCapacityFlowCSV(filePath string) ([]CapacityFlowRow, error) {
	records, err := ReadCSV(filePath)
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file %s has no data rows", filePath)
	}

	// Build header index: headerName -> column index
	header := records[0]
	colIndex := make(map[string]int, len(header))
	for i, h := range header {
		colIndex[strings.TrimSpace(h)] = i
	}

	var rows []CapacityFlowRow
	for lineNum, rec := range records[1:] {
		if len(rec) < 3 {
			continue
		}
		nodeID := strings.TrimSpace(rec[0])
		nodeType := strings.TrimSpace(rec[1])
		name := strings.TrimSpace(rec[2])
		if nodeID == "" || nodeType == "" {
			continue
		}

		mappings, ok := nodeTypeMappings[nodeType]
		if !ok {
			log.Printf("capacity_csv_parser: no mapping for node_type %q at line %d, skipping", nodeType, lineNum+2)
			continue
		}

		var vars []CapacityVariable
		for _, m := range mappings {
			idx, exists := colIndex[m.CSVHeader]
			if !exists {
				continue
			}
			if idx >= len(rec) {
				continue
			}
			raw := strings.TrimSpace(rec[idx])
			if raw == "" {
				continue
			}
			val, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				log.Printf("capacity_csv_parser: parse error for %s.%s at line %d: %v", nodeID, m.CSVHeader, lineNum+2, err)
				continue
			}
			vars = append(vars, CapacityVariable{
				VariableName: m.VariableName,
				Value:        val,
				Unit:         m.Unit,
			})
		}

		if len(vars) > 0 {
			rows = append(rows, CapacityFlowRow{
				NodeID:    nodeID,
				NodeType:  nodeType,
				Name:      name,
				Variables: vars,
			})
		}
	}

	return rows, nil
}

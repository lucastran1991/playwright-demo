package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
	"gorm.io/gorm"
)

// TraceResponse is the top-level response for dependency/impact tracing.
type TraceResponse struct {
	Source     SourceNode                    `json:"source"`
	Upstream   []TraceLevelGroup             `json:"upstream,omitempty"`
	Local      []TraceLocalGroup             `json:"local,omitempty"`
	Downstream []TraceLevelGroup             `json:"downstream,omitempty"`
	Load       []TraceLocalGroup             `json:"load,omitempty"`
	Capacity   map[string]map[string]float64 `json:"capacity,omitempty"` // nodeID -> varName -> value
}

// SourceNode identifies the node being traced.
type SourceNode struct {
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	NodeType string `json:"node_type"`
	Topology string `json:"topology"`
}

// TraceLevelGroup groups traced nodes by level and topology.
type TraceLevelGroup struct {
	Level    int                     `json:"level"`
	Topology string                  `json:"topology"`
	Nodes    []repository.TracedNode `json:"nodes"`
}

// TraceLocalGroup groups traced nodes by topology (no level).
type TraceLocalGroup struct {
	Topology string                  `json:"topology"`
	Nodes    []repository.TracedNode `json:"nodes"`
}

// DependencyTracer resolves actual node instances from type-level rules.
type DependencyTracer struct {
	repo           *repository.TracerRepository
	capRepo        *repository.CapacityRepository // optional, nil = no capacity enrichment
	topoLookup     map[string]string              // nodeType -> topology name
	slugLookup     map[string]string              // topology name -> blueprint_type slug
	topoLookupList []model.CapacityNodeType       // full list for IsCapacityNode checks
}

// SetCapacityRepo injects the capacity repository for /trace/full enrichment.
func (t *DependencyTracer) SetCapacityRepo(repo *repository.CapacityRepository) {
	t.capRepo = repo
}

// NewDependencyTracer creates a new DependencyTracer with DB-backed topology lookup.
func NewDependencyTracer(repo *repository.TracerRepository) *DependencyTracer {
	t := &DependencyTracer{repo: repo}
	t.RefreshLookups()
	return t
}

// RefreshLookups reloads topology mappings from capacity_node_types and blueprint_types tables.
func (t *DependencyTracer) RefreshLookups() {
	// Build nodeType -> topology mapping from capacity_node_types
	types, err := t.repo.ListCapacityNodeTypes()
	if err != nil {
		log.Printf("WARNING: failed to load capacity node types: %v", err)
		t.topoLookup = make(map[string]string)
	} else {
		lookup := make(map[string]string, len(types))
		for _, ct := range types {
			lookup[ct.NodeType] = ct.Topology
		}
		t.topoLookup = lookup
		t.topoLookupList = types
	}

	// Build topology name -> slug mapping from blueprint_types.
	// Uses case-insensitive prefix matching because CSV topology names
	// may differ from DB names (e.g. "Whitespace Blueprint" vs "Whitespace").
	t.slugLookup = make(map[string]string)
	btypes, err := t.repo.ListBlueprintTypes()
	if err != nil {
		log.Printf("WARNING: failed to load blueprint types: %v", err)
	} else {
		for _, bt := range btypes {
			t.slugLookup[bt.Name] = bt.Slug
		}
	}
	// Also build reverse: for each unique topology in capacity_node_types,
	// find the best matching blueprint_type by case-insensitive prefix.
	for _, ct := range types {
		topo := ct.Topology
		if _, ok := t.slugLookup[topo]; ok {
			continue // exact match exists
		}
		topoLower := strings.ToLower(topo)
		for _, bt := range btypes {
			if strings.HasPrefix(topoLower, strings.ToLower(bt.Name)) ||
				strings.HasPrefix(strings.ToLower(bt.Name), topoLower) {
				t.slugLookup[topo] = bt.Slug
				break
			}
		}
	}
}

// TraceDependencies finds all upstream and local dependencies for a node.
func (t *DependencyTracer) TraceDependencies(nodeID string, maxLevels int, includeLocal bool) (*TraceResponse, error) {
	node, err := t.repo.FindNodeByStringID(nodeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("node not found: %s", nodeID)
		}
		return nil, err
	}

	rules, err := t.repo.GetDependencyRules(node.NodeType)
	if err != nil {
		return nil, fmt.Errorf("get dependency rules: %w", err)
	}

	resp := &TraceResponse{
		Source: SourceNode{NodeID: node.NodeID, Name: node.Name, NodeType: node.NodeType, Topology: t.lookupTopology(node.NodeType)},
	}

	upstreamByTopo, localByTopo := t.groupDepRules(rules)

	for topo, allowedTypes := range upstreamByTopo {
		slug := t.resolveSlug(topo)
		if slug == "" {
			continue
		}

		// Try direct path first
		var filtered []repository.TracedNode
		nodes, err := t.repo.FindUpstreamNodes(node.ID, slug, maxLevels)
		if err == nil {
			filtered = filterByTypes(nodes, allowedTypes)
		}

		// If direct path found results, use them
		if len(filtered) > 0 {
			resp.Upstream = append(resp.Upstream, groupByLevel(filtered, topo)...)
			continue
		}

		// Spatial-bridge path: find bridge nodes via spatial hierarchy.
		// Try all bridge nodes, pick the one with the most upstream results.
		bridgeNodes, err := t.repo.FindBridgeNodesViaSpatial(node.ID, slug, 5)
		if err != nil || len(bridgeNodes) == 0 {
			continue
		}

		var bestFiltered []repository.TracedNode
		for _, bridge := range bridgeNodes {
			upNodes, err := t.repo.FindUpstreamNodes(bridge.ID, slug, maxLevels)
			if err != nil {
				continue
			}
			if allowedTypes[bridge.NodeType] {
				upNodes = append([]repository.TracedNode{
					{ID: bridge.ID, NodeID: bridge.NodeID, Name: bridge.Name,
						NodeType: bridge.NodeType, Level: bridge.Level,
						ParentNodeID: &node.NodeID},
				}, upNodes...)
			}
			for i := range upNodes {
				upNodes[i].Level += bridge.Level
			}
			candidate := filterByTypes(upNodes, allowedTypes)
			// Remove nodes whose shifted level exceeds maxLevels
			var trimmed []repository.TracedNode
			for _, n := range candidate {
				if n.Level <= maxLevels {
					trimmed = append(trimmed, n)
				}
			}
			if len(trimmed) > len(bestFiltered) {
				bestFiltered = trimmed
			}
		}
		if len(bestFiltered) > 0 {
			resp.Upstream = append(resp.Upstream, groupByLevel(bestFiltered, topo)...)
		}
	}

	if includeLocal {
		for topo, allowedTypes := range localByTopo {
			slug := t.resolveSlug(topo)
			if slug == "" {
				continue
			}
			nodes, err := t.repo.FindLocalNodes(node.ID, slug)
			if err != nil {
				log.Printf("WARNING: local trace failed for %s in %s: %v", nodeID, topo, err)
				continue
			}
			filtered := filterByTypes(nodes, allowedTypes)
			if len(filtered) > 0 {
				resp.Local = append(resp.Local, TraceLocalGroup{Topology: topo, Nodes: filtered})
				continue
			}

			// Bridge fallback for local: find whitespace/spatial children with target topology edges
			bridgeNodes, err := t.repo.FindBridgeNodesViaSpatial(node.ID, slug, 5)
			if err != nil || len(bridgeNodes) == 0 {
				continue
			}
			for _, bridge := range bridgeNodes {
				localNodes, err := t.repo.FindLocalNodes(bridge.ID, slug)
				if err != nil {
					continue
				}
				bridgeFiltered := filterByTypes(localNodes, allowedTypes)
				if len(bridgeFiltered) > 0 {
					resp.Local = append(resp.Local, TraceLocalGroup{Topology: topo, Nodes: bridgeFiltered})
					break
				}
			}
		}
	}

	return resp, nil
}

// TraceImpacts finds all downstream and load impacts for a node.
func (t *DependencyTracer) TraceImpacts(nodeID string, maxLevels int) (*TraceResponse, error) {
	node, err := t.repo.FindNodeByStringID(nodeID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("node not found: %s", nodeID)
		}
		return nil, err
	}

	rules, err := t.repo.GetImpactRules(node.NodeType)
	if err != nil {
		return nil, fmt.Errorf("get impact rules: %w", err)
	}

	resp := &TraceResponse{
		Source: SourceNode{NodeID: node.NodeID, Name: node.Name, NodeType: node.NodeType, Topology: t.lookupTopology(node.NodeType)},
	}

	downstreamByTopo, loadByTopo := t.groupImpactRules(rules)

	// Collect all load types into a single set for cross-topology lookup
	allLoadTypes := make(map[string]bool)
	for _, types := range loadByTopo {
		for nt := range types {
			allLoadTypes[nt] = true
		}
	}

	// Track infra topology slugs for downstream traversal.
	// Include source node's own topology so Load search has a starting point
	// even when there are no Downstream rules (e.g. Air Zone has only Load rules).
	infraSlugSet := make(map[string]bool)
	sourceTopo := t.lookupTopology(node.NodeType)
	if slug := t.resolveSlug(sourceTopo); slug != "" {
		infraSlugSet[slug] = true
	}

	for topo, allowedTypes := range downstreamByTopo {
		slug := t.resolveSlug(topo)
		if slug == "" {
			continue
		}
		infraSlugSet[slug] = true
		nodes, err := t.repo.FindDownstreamNodes(node.ID, slug, maxLevels)
		if err != nil {
			log.Printf("WARNING: downstream trace failed for %s in %s: %v", nodeID, topo, err)
			continue
		}
		filtered := filterByTypes(nodes, allowedTypes)
		resp.Downstream = append(resp.Downstream, groupByLevel(filtered, topo)...)
	}

	// Load impact only applies to Capacity Nodes (IsCapacityNode=true per spec).
	// Non-capacity nodes (Switch Gear, Rack PDU, ACU, CDU, etc.) only have Downstream impact.
	isCapacityNode := false
	for _, ct := range t.topoLookupList {
		if ct.NodeType == node.NodeType && ct.IsCapacityNode {
			isCapacityNode = true
			break
		}
	}

	// Load nodes (Rack, Row, etc.) only shown for Capacity Nodes per spec.
	// Connected via infra topologies. Three strategies run unconditionally and merge:
	// 1. Walk downstream in infra topologies with extended depth (finds Racks in Cooling)
	// 2. Find spatial ancestors of downstream nodes (Rack is parent of RACKPDU)
	// 3. Find spatial descendants of downstream nodes (Zone contains Rack as child)
	if isCapacityNode && len(allLoadTypes) > 0 {
		loadMaxLevels := 10
		loadGroupMap := make(map[string][]repository.TracedNode)
		seen := make(map[string]bool)

		// Collect all downstream DB IDs for spatial lookups.
		// Uses loadMaxLevels (10) instead of user maxLevels to reach deeper leaf nodes
		// like RackPDU that connect to spatial Load nodes (Rack, Row, Zone).
		var allDownstreamDBIDs []uint
		for slug := range infraSlugSet {
			nodes, err := t.repo.FindDownstreamNodes(node.ID, slug, loadMaxLevels)
			if err != nil {
				continue
			}
			for _, n := range nodes {
				allDownstreamDBIDs = append(allDownstreamDBIDs, n.ID)
				// Strategy 1: check if this downstream node is a load type
				if allLoadTypes[n.NodeType] && !seen[n.NodeID] {
					seen[n.NodeID] = true
					loadTopo := t.lookupTopology(n.NodeType)
					loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
				}
			}
		}

		loadTypeSlice := make([]string, 0, len(allLoadTypes))
		for nt := range allLoadTypes {
			loadTypeSlice = append(loadTypeSlice, nt)
		}

		if len(allDownstreamDBIDs) > 0 && len(loadTypeSlice) > 0 {
			// Strategy 2: spatial ancestors (Rack is parent of RackPDU in spatial)
			spatialLoads, err := t.repo.FindSpatialAncestorsOfType(allDownstreamDBIDs, loadTypeSlice)
			if err == nil {
				for _, n := range spatialLoads {
					if !seen[n.NodeID] {
						seen[n.NodeID] = true
						loadTopo := t.lookupTopology(n.NodeType)
						loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
					}
				}
			}

			// Strategy 3: spatial descendants (Zone contains Rack as child)
			spatialDescs, err := t.repo.FindSpatialDescendantsOfType(allDownstreamDBIDs, loadTypeSlice)
			if err == nil {
				for _, n := range spatialDescs {
					if !seen[n.NodeID] {
						seen[n.NodeID] = true
						loadTopo := t.lookupTopology(n.NodeType)
						loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
					}
				}
			}
		}

		for topo, nodes := range loadGroupMap {
			resp.Load = append(resp.Load, TraceLocalGroup{Topology: topo, Nodes: nodes})
		}
	}

	return resp, nil
}

// TraceFull returns combined dependency + impact trace in a single response.
func (t *DependencyTracer) TraceFull(nodeID string, maxLevels int) (*TraceResponse, error) {
	depResp, depErr := t.TraceDependencies(nodeID, maxLevels, true)
	impResp, impErr := t.TraceImpacts(nodeID, maxLevels)

	if depErr != nil && impErr != nil {
		return nil, depErr
	}
	if depErr != nil {
		log.Printf("WARNING: TraceFull partial failure (deps) for %s: %v", nodeID, depErr)
	}
	if impErr != nil {
		log.Printf("WARNING: TraceFull partial failure (impacts) for %s: %v", nodeID, impErr)
	}

	resp := depResp
	if resp == nil {
		resp = impResp
	}
	if resp == nil {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}

	if impResp != nil {
		resp.Downstream = impResp.Downstream
		resp.Load = impResp.Load
	}

	// Enrich with capacity data if capacity repo is available
	if t.capRepo != nil {
		allNodeIDs := collectAllNodeIDs(resp)
		capMap, err := t.capRepo.GetCapacityMapForNodes(allNodeIDs)
		if err == nil && len(capMap) > 0 {
			resp.Capacity = capMap
		}
	}

	return resp, nil
}

// collectAllNodeIDs gathers unique node IDs from all parts of a TraceResponse.
func collectAllNodeIDs(resp *TraceResponse) []string {
	seen := map[string]bool{resp.Source.NodeID: true}
	for _, groups := range [][]TraceLevelGroup{resp.Upstream, resp.Downstream} {
		for _, g := range groups {
			for _, n := range g.Nodes {
				seen[n.NodeID] = true
			}
		}
	}
	for _, groups := range [][]TraceLocalGroup{resp.Local, resp.Load} {
		for _, g := range groups {
			for _, n := range g.Nodes {
				seen[n.NodeID] = true
			}
		}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

// resolveSlug maps a topology name to its blueprint_type slug using cached DB data.
func (t *DependencyTracer) resolveSlug(topology string) string {
	if slug, ok := t.slugLookup[topology]; ok {
		return slug
	}
	log.Printf("WARNING: no blueprint_type slug found for topology %q", topology)
	return ""
}

// lookupTopology returns the topology for a node type from cached DB data.
func (t *DependencyTracer) lookupTopology(nodeType string) string {
	if topo, ok := t.topoLookup[nodeType]; ok {
		return topo
	}
	return "Electrical System" // fallback
}

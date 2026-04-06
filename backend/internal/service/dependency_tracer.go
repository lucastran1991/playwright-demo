package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/user/app/internal/repository"
	"gorm.io/gorm"
)

// TraceResponse is the top-level response for dependency/impact tracing.
type TraceResponse struct {
	Source     SourceNode        `json:"source"`
	Upstream   []TraceLevelGroup `json:"upstream,omitempty"`
	Local      []TraceLocalGroup `json:"local,omitempty"`
	Downstream []TraceLevelGroup `json:"downstream,omitempty"`
	Load       []TraceLocalGroup `json:"load,omitempty"`
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
	repo       *repository.TracerRepository
	topoLookup map[string]string // nodeType -> topology name
	slugLookup map[string]string // topology name -> blueprint_type slug
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
		nodes, err := t.repo.FindUpstreamNodes(node.ID, slug, maxLevels)
		if err != nil {
			log.Printf("WARNING: upstream trace failed for %s in %s: %v", nodeID, topo, err)
			continue
		}
		filtered := filterByTypes(nodes, allowedTypes)

		// If no matching upstream found, also trace from this node's direct children
		// in the same topology. This handles cases like Rack→RACKPDU in electrical
		// where RPP feeds RACKPDU (not Rack directly).
		if len(filtered) == 0 {
			children, err := t.repo.FindDownstreamNodes(node.ID, slug, 1)
			if err == nil {
				for _, child := range children {
					childUpstream, err := t.repo.FindUpstreamNodes(child.ID, slug, maxLevels)
					if err != nil {
						continue
					}
					// Shift levels +1 since we went through a child first
					for i := range childUpstream {
						childUpstream[i].Level++
					}
					filtered = append(filtered, filterByTypes(childUpstream, allowedTypes)...)
				}
			}
		}

		resp.Upstream = append(resp.Upstream, groupByLevel(filtered, topo)...)
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

	// Load nodes (Rack, Row, etc.) live as children in infra topologies
	// (e.g. AIRZONE → RACK in Cooling), not in their own topology (Spatial).
	// Two strategies:
	// 1. Walk downstream in infra topologies with extended depth (works for Cooling)
	// 2. Find spatial ancestors of downstream nodes (works for Electrical)
	if len(allLoadTypes) > 0 {
		loadMaxLevels := 10
		loadGroupMap := make(map[string][]repository.TracedNode)
		seen := make(map[string]bool)

		// Strategy 1: walk downstream in infra topologies (finds Racks in Cooling edges)
		for slug := range infraSlugSet {
			nodes, err := t.repo.FindDownstreamNodes(node.ID, slug, loadMaxLevels)
			if err != nil {
				continue
			}
			for _, n := range nodes {
				if allLoadTypes[n.NodeType] && !seen[n.NodeID] {
					seen[n.NodeID] = true
					loadTopo := t.lookupTopology(n.NodeType)
					loadGroupMap[loadTopo] = append(loadGroupMap[loadTopo], n)
				}
			}
		}

		// Strategy 2: find spatial ancestors of deep downstream nodes
		// (Rack is spatial parent of RACKPDU, which is deep in electrical chain).
		// Walk deeper than resp.Downstream (which uses maxLevels) to reach leaf nodes.
		if len(loadGroupMap) == 0 {
			loadTypeSlice := make([]string, 0, len(allLoadTypes))
			for nt := range allLoadTypes {
				loadTypeSlice = append(loadTypeSlice, nt)
			}
			var allDownstreamDBIDs []uint
			for slug := range infraSlugSet {
				deepNodes, err := t.repo.FindDownstreamNodes(node.ID, slug, loadMaxLevels)
				if err != nil {
					continue
				}
				for _, n := range deepNodes {
					allDownstreamDBIDs = append(allDownstreamDBIDs, n.ID)
				}
			}
			if len(allDownstreamDBIDs) > 0 {
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
			}
		}

		for topo, nodes := range loadGroupMap {
			resp.Load = append(resp.Load, TraceLocalGroup{Topology: topo, Nodes: nodes})
		}
	}

	return resp, nil
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

package service

import (
	"fmt"
	"log"

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

	// Build topology name -> slug mapping from blueprint_types
	t.slugLookup = make(map[string]string)
	btypes, err := t.repo.ListBlueprintTypes()
	if err != nil {
		log.Printf("WARNING: failed to load blueprint types: %v", err)
	} else {
		for _, bt := range btypes {
			t.slugLookup[bt.Name] = bt.Slug
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
		Source: SourceNode{NodeID: node.NodeID, Name: node.Name, NodeType: node.NodeType},
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
		Source: SourceNode{NodeID: node.NodeID, Name: node.Name, NodeType: node.NodeType},
	}

	downstreamByTopo, loadByTopo := t.groupImpactRules(rules)

	for topo, allowedTypes := range downstreamByTopo {
		slug := t.resolveSlug(topo)
		if slug == "" {
			continue
		}
		nodes, err := t.repo.FindDownstreamNodes(node.ID, slug, maxLevels)
		if err != nil {
			log.Printf("WARNING: downstream trace failed for %s in %s: %v", nodeID, topo, err)
			continue
		}
		filtered := filterByTypes(nodes, allowedTypes)
		resp.Downstream = append(resp.Downstream, groupByLevel(filtered, topo)...)
	}

	for topo, allowedTypes := range loadByTopo {
		slug := t.resolveSlug(topo)
		if slug == "" {
			continue
		}
		nodes, err := t.repo.FindDownstreamNodes(node.ID, slug, maxLevels)
		if err != nil {
			log.Printf("WARNING: load trace failed for %s in %s: %v", nodeID, topo, err)
			continue
		}
		filtered := filterByTypes(nodes, allowedTypes)
		if len(filtered) > 0 {
			resp.Load = append(resp.Load, TraceLocalGroup{Topology: topo, Nodes: filtered})
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

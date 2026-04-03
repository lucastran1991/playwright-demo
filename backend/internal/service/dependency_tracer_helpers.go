package service

import (
	"github.com/user/app/internal/model"
	"github.com/user/app/internal/repository"
)

// groupDepRules separates dependency rules into upstream and local, grouped by topology.
func (t *DependencyTracer) groupDepRules(rules []model.DependencyRule) (upstream, local map[string]map[string]bool) {
	upstream = make(map[string]map[string]bool)
	local = make(map[string]map[string]bool)
	for _, r := range rules {
		topo := t.lookupTopology(r.DependencyNodeType)
		target := upstream
		if r.TopologicalRelationship == "Local" {
			target = local
		}
		if target[topo] == nil {
			target[topo] = make(map[string]bool)
		}
		target[topo][r.DependencyNodeType] = true
	}
	return upstream, local
}

// groupImpactRules separates impact rules into downstream and load, grouped by topology.
func (t *DependencyTracer) groupImpactRules(rules []model.ImpactRule) (downstream, load map[string]map[string]bool) {
	downstream = make(map[string]map[string]bool)
	load = make(map[string]map[string]bool)
	for _, r := range rules {
		topo := t.lookupTopology(r.ImpactNodeType)
		target := downstream
		if r.TopologicalRelationship == "Load" {
			target = load
		}
		if target[topo] == nil {
			target[topo] = make(map[string]bool)
		}
		target[topo][r.ImpactNodeType] = true
	}
	return downstream, load
}

// filterByTypes keeps only nodes whose type is in the allowed set.
func filterByTypes(nodes []repository.TracedNode, allowed map[string]bool) []repository.TracedNode {
	if len(allowed) == 0 {
		return nodes
	}
	var filtered []repository.TracedNode
	for _, n := range nodes {
		if allowed[n.NodeType] {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// groupByLevel groups traced nodes by their level, returning one group per level.
func groupByLevel(nodes []repository.TracedNode, topology string) []TraceLevelGroup {
	levelMap := make(map[int][]repository.TracedNode)
	for _, n := range nodes {
		levelMap[n.Level] = append(levelMap[n.Level], n)
	}
	var groups []TraceLevelGroup
	for level, ns := range levelMap {
		groups = append(groups, TraceLevelGroup{Level: level, Topology: topology, Nodes: ns})
	}
	return groups
}

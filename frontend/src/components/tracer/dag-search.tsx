'use client'

import { useState, useEffect, useRef } from "react"
import { useQuery } from "@tanstack/react-query"
import { Search, X, Loader2, Filter } from "lucide-react"
import { apiFetch } from "@/lib/api-client"
import type { SearchNode } from "./dag-types"

interface SearchResponse {
  data: SearchNode[]
  total: number
}

interface DAGSearchProps {
  onSelect: (nodeId: string) => void
  onClear: () => void
}

// Common node types for quick filtering
const NODE_TYPES = [
  "All Types",
  "Rack", "RPP", "Room PDU", "UPS", "BESS", "Switch Gear",
  "Utility Feed", "Generator", "Air Zone", "Air Cooling Unit",
  "CDU", "Cooling Distribution", "Cooling Plant", "Liquid Loop",
  "RDHx", "DTC", "Capacity Cell", "Room Bundle",
]

export default function DAGSearch({ onSelect, onClear }: DAGSearchProps) {
  const [inputValue, setInputValue] = useState("")
  const [debouncedQuery, setDebouncedQuery] = useState("")
  const [isOpen, setIsOpen] = useState(false)
  const [selectedLabel, setSelectedLabel] = useState("")
  const [typeFilter, setTypeFilter] = useState("All Types")
  const [showTypeDropdown, setShowTypeDropdown] = useState(false)
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedQuery(inputValue), 300)
    return () => clearTimeout(timer)
  }, [inputValue])

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
        setShowTypeDropdown(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const { data, isFetching } = useQuery({
    queryKey: ["node-search", debouncedQuery],
    queryFn: () =>
      apiFetch<SearchResponse>(`/api/blueprints/nodes?search=${encodeURIComponent(debouncedQuery)}&limit=50`),
    enabled: debouncedQuery.length >= 2,
    staleTime: 30_000,
  })

  // Client-side type filter on results
  const allResults = data?.data ?? []
  const results = typeFilter === "All Types"
    ? allResults.slice(0, 20)
    : allResults.filter((n) => n.node_type === typeFilter).slice(0, 20)

  function handleSelect(node: SearchNode) {
    setSelectedLabel(`${node.node_id} — ${node.name}`)
    setInputValue("")
    setDebouncedQuery("")
    setIsOpen(false)
    onSelect(node.node_id)
  }

  function handleClear() {
    setInputValue("")
    setDebouncedQuery("")
    setSelectedLabel("")
    setIsOpen(false)
    setTypeFilter("All Types")
    onClear()
  }

  const showDropdown = isOpen && debouncedQuery.length >= 2

  return (
    <div ref={containerRef} className="absolute top-4 left-1/2 -translate-x-1/2 z-10 w-[480px]">
      <div className="flex gap-0">
        {/* Search input */}
        <div className="relative flex items-center flex-1 rounded-l-lg border border-r-0 border-border bg-card shadow-lg">
          <Search className="absolute left-3 h-4 w-4 text-muted-foreground pointer-events-none" />
          {selectedLabel && !inputValue ? (
            <div className="flex-1 px-10 py-2.5 text-sm truncate text-foreground">{selectedLabel}</div>
          ) : (
            <input
              type="text"
              value={inputValue}
              onChange={(e) => { setInputValue(e.target.value); setIsOpen(true) }}
              onFocus={() => setIsOpen(true)}
              placeholder="Search nodes..."
              className="flex-1 bg-transparent px-10 py-2.5 text-sm outline-none placeholder:text-muted-foreground"
            />
          )}
          <div className="absolute right-3 flex items-center">
            {isFetching ? (
              <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
            ) : (selectedLabel || inputValue) ? (
              <button onClick={handleClear} className="text-muted-foreground hover:text-foreground transition-colors">
                <X className="h-4 w-4" />
              </button>
            ) : null}
          </div>
        </div>

        {/* Type filter button */}
        <button
          onClick={() => setShowTypeDropdown((v) => !v)}
          className="flex items-center gap-1.5 px-3 py-2.5 rounded-r-lg border border-border bg-card shadow-lg text-xs font-medium text-muted-foreground hover:text-foreground transition-colors whitespace-nowrap"
        >
          <Filter className="h-3.5 w-3.5" />
          <span className="max-w-[80px] truncate">{typeFilter === "All Types" ? "Type" : typeFilter}</span>
        </button>
      </div>

      {/* Type dropdown */}
      {showTypeDropdown && (
        <div className="absolute top-full right-0 mt-1 w-[180px] rounded-lg border border-border bg-card shadow-xl overflow-hidden max-h-64 overflow-y-auto">
          {NODE_TYPES.map((t) => (
            <button
              key={t}
              onClick={() => { setTypeFilter(t); setShowTypeDropdown(false) }}
              className={`w-full text-left px-3 py-2 text-xs transition-colors border-b border-border/30 last:border-0 ${
                typeFilter === t ? "bg-primary/10 text-primary font-medium" : "hover:bg-accent text-foreground"
              }`}
            >
              {t}
            </button>
          ))}
        </div>
      )}

      {/* Search results dropdown */}
      {showDropdown && (
        <div className="mt-1 rounded-lg border border-border bg-card shadow-xl overflow-hidden max-h-64 overflow-y-auto">
          {results.length === 0 && !isFetching ? (
            <div className="px-4 py-3 text-sm text-muted-foreground">
              No {typeFilter !== "All Types" ? `${typeFilter} ` : ""}nodes found
            </div>
          ) : (
            results.map((node) => (
              <button
                key={node.id}
                onClick={() => handleSelect(node)}
                className="w-full flex items-center gap-3 px-4 py-2.5 text-left hover:bg-accent transition-colors border-b border-border/50 last:border-0"
              >
                <span className="font-mono text-xs text-primary shrink-0">{node.node_id}</span>
                <span className="text-sm text-foreground truncate flex-1">{node.name}</span>
                <span className="shrink-0 rounded-full bg-muted px-2 py-0.5 text-[10px] text-muted-foreground uppercase tracking-wide">
                  {node.node_type}
                </span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  )
}

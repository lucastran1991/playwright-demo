'use client'

import { useState, useEffect, useRef } from "react"
import { useQuery } from "@tanstack/react-query"
import { Search, X, Loader2 } from "lucide-react"
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

export default function DAGSearch({ onSelect, onClear }: DAGSearchProps) {
  const [inputValue, setInputValue] = useState("")
  const [debouncedQuery, setDebouncedQuery] = useState("")
  const [isOpen, setIsOpen] = useState(false)
  const [selectedLabel, setSelectedLabel] = useState("")
  const containerRef = useRef<HTMLDivElement>(null)

  // Debounce input by 300ms
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedQuery(inputValue)
    }, 300)
    return () => clearTimeout(timer)
  }, [inputValue])

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setIsOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  const { data, isFetching } = useQuery({
    queryKey: ["node-search", debouncedQuery],
    queryFn: () =>
      apiFetch<SearchResponse>(`/api/blueprints/nodes?search=${encodeURIComponent(debouncedQuery)}&limit=20`),
    enabled: debouncedQuery.length >= 2,
    staleTime: 30_000,
  })

  const results = data?.data ?? []

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
    onClear()
  }

  const showDropdown = isOpen && debouncedQuery.length >= 2

  return (
    <div
      ref={containerRef}
      className="absolute top-4 left-1/2 -translate-x-1/2 z-10 w-[400px]"
    >
      <div className="relative flex items-center rounded-lg border border-border bg-card shadow-lg">
        <Search className="absolute left-3 h-4 w-4 text-muted-foreground pointer-events-none" />

        {selectedLabel && !inputValue ? (
          <div className="flex-1 px-10 py-2.5 text-sm truncate text-foreground">
            {selectedLabel}
          </div>
        ) : (
          <input
            type="text"
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value)
              setIsOpen(true)
            }}
            onFocus={() => setIsOpen(true)}
            placeholder="Search for a node..."
            className="flex-1 bg-transparent px-10 py-2.5 text-sm outline-none placeholder:text-muted-foreground"
          />
        )}

        {/* Right icon: spinner or clear */}
        <div className="absolute right-3 flex items-center">
          {isFetching ? (
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          ) : (selectedLabel || inputValue) ? (
            <button
              onClick={handleClear}
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          ) : null}
        </div>
      </div>

      {/* Dropdown results */}
      {showDropdown && (
        <div className="mt-1 rounded-lg border border-border bg-card shadow-xl overflow-hidden max-h-64 overflow-y-auto">
          {results.length === 0 && !isFetching ? (
            <div className="px-4 py-3 text-sm text-muted-foreground">
              No nodes found for &quot;{debouncedQuery}&quot;
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

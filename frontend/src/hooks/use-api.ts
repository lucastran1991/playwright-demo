"use client"

import { useQuery } from "@tanstack/react-query"
import { useAuth } from "./use-auth"
import { apiFetch } from "@/lib/api-client"
import type { User } from "@/types"

// Centralized query keys for cache management
export const queryKeys = {
  user: {
    all: ["user"] as const,
    me: () => [...queryKeys.user.all, "me"] as const,
  },
}

// Fetch the currently authenticated user from the backend
export function useCurrentUser() {
  const { accessToken } = useAuth()

  return useQuery({
    queryKey: queryKeys.user.me(),
    queryFn: () =>
      apiFetch<{ data: User }>("/api/auth/me", {
        headers: { Authorization: `Bearer ${accessToken}` },
      }).then((res) => res.data),
    enabled: !!accessToken,
  })
}

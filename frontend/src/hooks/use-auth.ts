"use client"

import { useSession } from "next-auth/react"

export function useAuth() {
  const { data: session, status } = useSession()

  return {
    user: session?.user,
    accessToken: session?.accessToken,
    isAuthenticated: status === "authenticated",
    isLoading: status === "loading",
  }
}

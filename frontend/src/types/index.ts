// Core domain types matching the Go backend JSON responses

export interface User {
  id: number
  name: string
  email: string
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  user: User
}

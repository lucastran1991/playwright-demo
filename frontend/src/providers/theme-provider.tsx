"use client"

import { ThemeProvider as NextThemesProvider } from "next-themes"

// Wraps app with next-themes for class-based dark mode switching
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  return (
    <NextThemesProvider attribute="class" defaultTheme="system" enableSystem>
      {children}
    </NextThemesProvider>
  )
}

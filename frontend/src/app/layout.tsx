import type { Metadata } from "next"
import { DM_Sans, JetBrains_Mono } from "next/font/google"
import { ThemeProvider } from "@/providers/theme-provider"
import { QueryProvider } from "@/providers/query-provider"
import { SessionProvider } from "@/providers/session-provider"
import "./globals.css"

// Primary UI font — geometric, supports Vietnamese
const dmSans = DM_Sans({
  variable: "--font-sans",
  subsets: ["latin", "latin-ext"],
})

// Monospace font for code/data display
const jetbrainsMono = JetBrains_Mono({
  variable: "--font-mono",
  subsets: ["latin"],
})

export const metadata: Metadata = {
  title: "Playwright Demo",
  description: "Fullstack Go + Next.js scaffold",
}

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html
      lang="en"
      className={`${dmSans.variable} ${jetbrainsMono.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col">
        <SessionProvider>
          <ThemeProvider>
            <QueryProvider>{children}</QueryProvider>
          </ThemeProvider>
        </SessionProvider>
      </body>
    </html>
  )
}

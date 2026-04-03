import { ThemeToggle } from "@/components/dashboard/theme-toggle"

export default function TracerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      {/* Minimal header: just dark mode toggle */}
      <header className="sticky top-0 z-10 flex h-12 items-center justify-end border-b bg-background/80 backdrop-blur-sm px-4">
        <ThemeToggle />
      </header>
      <main className="p-4">{children}</main>
    </div>
  )
}

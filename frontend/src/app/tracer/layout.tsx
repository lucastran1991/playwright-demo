export default function TracerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="h-screen bg-background">
      {children}
    </div>
  )
}

export default function TracerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="h-screen h-dvh bg-background">
      {children}
    </div>
  )
}

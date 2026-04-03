import DependencyImpactDAG from '@/components/tracer/dependency-impact-dag'

export const metadata = { title: 'Dependency Tracer' }

export default function TracerPage() {
  return (
    <div className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold">Dependency Tracer</h1>
        <p className="text-muted-foreground text-sm">
          Search for a node to visualize upstream dependencies and downstream impacts
        </p>
      </div>
      <DependencyImpactDAG />
    </div>
  )
}

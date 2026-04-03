import DependencyImpactDAG from "@/components/tracer/dependency-impact-dag"

export const metadata = {
  title: "Dependency Tracer",
}

export default function TracerPage() {
  return <DependencyImpactDAG />
}

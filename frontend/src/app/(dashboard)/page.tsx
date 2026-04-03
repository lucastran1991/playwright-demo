"use client"

import { useCurrentUser } from "@/hooks/use-api"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"

const stats = [
  { title: "Total Users", value: "128" },
  { title: "Active Sessions", value: "42" },
  { title: "Uptime", value: "99.9%" },
]

export default function DashboardPage() {
  const { data: user, isLoading } = useCurrentUser()

  return (
    <div className="space-y-6">
      {isLoading ? (
        <Skeleton className="h-9 w-64" />
      ) : (
        <h1 className="text-3xl font-bold">
          Welcome back, {user?.name ?? "User"}
        </h1>
      )}

      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {stats.map((stat) => (
          <Card key={stat.title}>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {stat.title}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{stat.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

import { useLeaderboard } from '@/features/players/hooks'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/shared/ui/table'
import { Badge } from '@/shared/ui/badge'
import { Trophy, Medal } from 'lucide-react'
import { useState } from 'react'
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/shared/ui/select'
import { Label } from '@/shared/ui/label'

export function LeaderboardPage() {
  const [clubId, setClubId] = useState(1)
  const [metric, setMetric] = useState<'profit' | 'roi' | 'winrate'>('profit')

  const { data: leaderboard, isLoading } = useLeaderboard({
    club_id: clubId,
    metric,
    period: 'all',
  })

  const getRankBadge = (rank: number) => {
    switch (rank) {
      case 1:
        return <Badge variant="warning"><Trophy className="mr-1 h-3 w-3" /> 1</Badge>
      case 2:
        return <Badge variant="secondary"><Medal className="mr-1 h-3 w-3" /> 2</Badge>
      case 3:
        return <Badge variant="secondary"><Medal className="mr-1 h-3 w-3" /> 3</Badge>
      default:
        return <Badge variant="outline">#{rank}</Badge>
    }
  }

  const getMetricLabel = (metric: string) => {
    switch (metric) {
      case 'profit':
        return 'Profit'
      case 'roi':
        return 'ROI'
      case 'winrate':
        return 'Win Rate'
      default:
        return metric
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    )
  }

  const entries = leaderboard?.entries || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Рейтинг</h1>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Label htmlFor="club_id">Клуб ID:</Label>
            <input
              id="club_id"
              type="number"
              value={clubId}
              onChange={(e) => setClubId(parseInt(e.target.value) || 1)}
              className="w-20 p-2 border rounded-md"
            />
          </div>
          <div className="flex items-center gap-2">
            <Label htmlFor="metric">Метрика:</Label>
            <Select value={metric} onValueChange={(v) => setMetric(v as 'profit' | 'roi' | 'winrate')}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="profit">Profit</SelectItem>
                <SelectItem value="roi">ROI</SelectItem>
                <SelectItem value="winrate">Win Rate</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Лидерборд</CardTitle>
          <CardDescription>
            {leaderboard ? `${leaderboard.club_name} • ${getMetricLabel(leaderboard.metric)}` : 'Рейтинг игроков'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {entries.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Место</TableHead>
                  <TableHead>Игрок</TableHead>
                  <TableHead>Игр</TableHead>
                  <TableHead>{getMetricLabel(leaderboard?.metric || 'profit')}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {entries.map((entry, index) => (
                  <TableRow key={entry.player_id} className={index < 3 ? 'bg-muted/50' : ''}>
                    <TableCell>{getRankBadge(index + 1)}</TableCell>
                    <TableCell>
                      <div>
                        <p className="font-medium">{entry.player_name}</p>
                      </div>
                    </TableCell>
                    <TableCell>{entry.games_count}</TableCell>
                    <TableCell>
                      <Badge variant={entry.metric_value >= 0 ? 'success' : 'destructive'}>
                        {entry.metric_value >= 0 ? '+' : ''}{entry.metric_value.toFixed(2)}
                        {leaderboard?.metric === 'roi' || leaderboard?.metric === 'winrate' ? '%' : ' ₽'}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="flex flex-col items-center justify-center py-8 text-muted-foreground">
              <Trophy className="h-12 w-12 mb-4" />
              <p>Рейтинг пока не доступен</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

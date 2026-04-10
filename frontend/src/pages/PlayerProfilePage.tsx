import { useParams } from 'react-router-dom'
import { usePlayerStats } from '@/features/players/hooks'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { TrendingUp, TrendingDown } from 'lucide-react'

export function PlayerProfilePage() {
  const { id } = useParams<{ id: string }>()
  const playerId = parseInt(id || '0')
  const { data: stats, isLoading } = usePlayerStats(playerId)

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    )
  }

  if (!stats) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Игрок не найден</p>
      </div>
    )
  }

  const profitColor = stats.profit >= 0 ? 'text-green-600' : 'text-red-600'
  const roiColor = stats.roi >= 0 ? 'text-green-600' : 'text-red-600'

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Профиль игрока</h1>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Статистика игрока</CardTitle>
          <CardDescription>Подробная информация о результатах</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Всего игр</p>
              <p className="text-2xl font-bold">{stats.total_games}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Всего вложено</p>
              <p className="text-2xl font-bold">{stats.total_invested.toFixed(2)} ₽</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">Чипы</p>
              <p className="text-2xl font-bold">{stats.total_chips.toFixed(2)}</p>
            </div>
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">ITM</p>
              <p className="text-2xl font-bold">{stats.itm.toFixed(1)}%</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Profit</CardTitle>
            <CardDescription>Общая прибыль/убыток</CardDescription>
          </CardHeader>
          <CardContent>
            <div className={`text-4xl font-bold ${profitColor}`}>
              {stats.profit >= 0 ? '+' : ''}{stats.profit.toFixed(2)} ₽
            </div>
            <div className="flex items-center mt-2 text-sm text-muted-foreground">
              {stats.profit >= 0 ? (
                <TrendingUp className="h-4 w-4 mr-1 text-green-600" />
              ) : (
                <TrendingDown className="h-4 w-4 mr-1 text-red-600" />
              )}
              <span>ROI: <span className={roiColor}>{stats.roi.toFixed(1)}%</span></span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Инвестиции</CardTitle>
            <CardDescription>Расходы на buy-in и rebuy</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Buy-in:</span>
                <span className="font-medium">{stats.total_buy_in.toFixed(2)} ₽</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Rebuy:</span>
                <span className="font-medium">{stats.total_rebuy.toFixed(2)} ₽</span>
              </div>
              <div className="flex justify-between items-center pt-2 border-t">
                <span className="font-medium">Итого:</span>
                <span className="font-bold">{stats.total_invested.toFixed(2)} ₽</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

import { useAuthStore } from '@/features/auth/store'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Button } from '@/shared/ui/button'
import { useGames } from '@/features/games/hooks'
import { GameType, MoneyModel } from '@/entities/game/types'
import { Link } from 'react-router-dom'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'
import { useState } from 'react'

const formatDate = (dateString: string) => {
  try {
    return format(new Date(dateString), 'dd MMM yyyy, HH:mm', { locale: ru })
  } catch {
    return dateString
  }
}

const getGameTypeLabel = (type: GameType) => {
  switch (type) {
    case 'cash_time':
      return 'Cash Game (Time)'
    case 'cash_open':
      return 'Cash Game (Open)'
    case 'tournament':
      return 'Турнир'
    default:
      return type
  }
}

const getMoneyModelLabel = (model: MoneyModel) => {
  return model === 'real' ? 'Реальные деньги' : 'Чипы'
}

export function DashboardPage() {
  const user = useAuthStore((state) => state.user)
  const [clubId] = useState(1) // TODO: получить club_id из user или селектора клуба

  const { data, isLoading, error } = useGames({
    club_id: clubId,
    status: '',
    limit: 50,
    offset: 0,
  })

  const games = data?.games || []
  const activeGames = games.filter((g) => !g.end_time)
  const finishedGames = games.filter((g) => g.end_time)

  if (error) {
    console.error('Ошибка загрузки игр:', error)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">
          Добро пожаловать, {user?.first_name}!
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Всего игр</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{games.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Активные игры</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeGames.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Завершенные</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{finishedGames.length}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Ваша роль</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold capitalize">{user?.role}</div>
          </CardContent>
        </Card>
      </div>

      {/* Active Games */}
      <Card>
        <CardHeader>
          <CardTitle>Активные игры</CardTitle>
          <CardDescription>Текущие игры, в которых вы участвуете</CardDescription>
        </CardHeader>
        <CardContent>
          {activeGames.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">
              Нет активных игр
            </p>
          ) : (
            <div className="space-y-4">
              {activeGames.map((game) => (
                <div
                  key={game.id}
                  className="flex items-center justify-between p-4 border rounded-lg"
                >
                  <div className="space-y-1">
                    <p className="font-medium">
                      Игра #{game.id} - {getGameTypeLabel(game.type)}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {getMoneyModelLabel(game.money_model)} • Buy-in: {game.buy_in_amount} ₽
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Начало: {formatDate(game.start_time)}
                    </p>
                  </div>
                  <Button asChild>
                    <Link to={`/games/${game.id}`}>Подробности</Link>
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Recent Games */}
      <Card>
        <CardHeader>
          <CardTitle>Последние игры</CardTitle>
          <CardDescription>История завершенных игр</CardDescription>
        </CardHeader>
        <CardContent>
          {finishedGames.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">
              Нет завершенных игр
            </p>
          ) : (
            <div className="space-y-4">
              {finishedGames.slice(0, 5).map((game) => (
                <div
                  key={game.id}
                  className="flex items-center justify-between p-4 border rounded-lg"
                >
                  <div className="space-y-1">
                    <p className="font-medium">
                      Игра #{game.id} - {getGameTypeLabel(game.type)}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {formatDate(game.start_time)} - {formatDate(game.end_time!)}
                    </p>
                  </div>
                  <Button variant="outline" asChild>
                    <Link to={`/games/${game.id}`}>Результаты</Link>
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

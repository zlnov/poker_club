import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useGames, useCreateGame } from '@/features/games/hooks'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/shared/ui/table'
import { Badge } from '@/shared/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '@/shared/ui/select'
import { GameListItem, GameType, MoneyModel } from '@/entities/game/types'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'
import { Plus, Eye } from 'lucide-react'

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
      return 'Cash (Time)'
    case 'cash_open':
      return 'Cash (Open)'
    case 'tournament':
      return 'Турнир'
    default:
      return type
  }
}

const getStatusBadge = (game: GameListItem) => {
  if (game.end_time) {
    return <Badge variant="secondary">Завершена</Badge>
  }
  return <Badge variant="success">Активна</Badge>
}

export function GamesPage() {
  const [clubId, setClubId] = useState(1)
  const [status, setStatus] = useState<'active' | 'finished' | ''>('')
  const [limit, setLimit] = useState(50)
  const [offset, setOffset] = useState(0)

  const { data, isLoading, error } = useGames({
    club_id: clubId,
    status,
    limit,
    offset,
  })

  const games = data?.games || []
  const total = data?.total || 0

  if (error) {
    console.error('Ошибка загрузки игр:', error)
  }

  const createGameMutation = useCreateGame()
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [formData, setFormData] = useState({
    club_id: 1,
    banker_id: 1,
    type: 'cash_time' as GameType,
    money_model: 'real' as MoneyModel,
    buy_in_amount: 0,
    rebuy_allowed: false,
    rebuy_amount: 0,
    max_rebuys_per_player: 0,
    start_time: '',
    min_players: 2,
    max_players: 10,
    ranking_primary: 'chips',
    ranking_secondary: '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createGameMutation.mutate({
      ...formData,
      start_time: new Date(formData.start_time).toISOString(),
      rebuy_amount: formData.rebuy_allowed ? formData.rebuy_amount : null,
      max_rebuys_per_player: formData.rebuy_allowed ? formData.max_rebuys_per_player : null,
      duration: null,
      ranking_secondary: formData.ranking_secondary || null,
    })
    setIsCreateDialogOpen(false)
    setFormData({
      club_id: 1,
      banker_id: 1,
      type: 'cash_time',
      money_model: 'real',
      buy_in_amount: 0,
      rebuy_allowed: false,
      rebuy_amount: 0,
      max_rebuys_per_player: 0,
      start_time: '',
      min_players: 2,
      max_players: 10,
      ranking_primary: 'chips',
      ranking_secondary: '',
    })
  }

  const handleStatusChange = (value: string) => {
    setStatus(value as 'active' | 'finished' | '')
    setOffset(0) // Reset pagination on filter change
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
        <h1 className="text-3xl font-bold">Игры</h1>
        <Button onClick={() => setIsCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Создать игру
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Все игры</CardTitle>
          <CardDescription>
            Список всех игр в клубе • Всего: {total} • Отобрано: {games.length}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Filters */}
          <div className="flex gap-4 mb-4">
            <div className="flex items-center gap-2">
              <Label htmlFor="club_id">Клуб ID:</Label>
              <Input
                id="club_id"
                type="number"
                value={clubId}
                onChange={(e) => setClubId(parseInt(e.target.value) || 1)}
                className="w-24"
              />
            </div>
            <div className="flex items-center gap-2">
              <Label htmlFor="status">Статус:</Label>
              <Select value={status} onValueChange={handleStatusChange}>
                <SelectTrigger className="w-32">
                  <SelectValue placeholder="Все" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">Все</SelectItem>
                  <SelectItem value="active">Активные</SelectItem>
                  <SelectItem value="finished">Завершенные</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2">
              <Label htmlFor="limit">На странице:</Label>
              <Select value={limit.toString()} onValueChange={(v) => { setLimit(parseInt(v)); setOffset(0); }}>
                <SelectTrigger className="w-20">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="25">25</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {games.length > 0 ? (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>ID</TableHead>
                    <TableHead>Тип</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Начало</TableHead>
                    <TableHead>Buy-in</TableHead>
                    <TableHead>Участников</TableHead>
                    <TableHead>Действия</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {games.map((game: GameListItem) => (
                    <TableRow key={game.id}>
                      <TableCell>#{game.id}</TableCell>
                      <TableCell>{getGameTypeLabel(game.type)}</TableCell>
                      <TableCell>{getStatusBadge(game)}</TableCell>
                      <TableCell>{formatDate(game.start_time)}</TableCell>
                      <TableCell>{game.buy_in_amount} ₽</TableCell>
                      <TableCell>{game.participants_count}</TableCell>
                      <TableCell>
                        <Button variant="outline" size="sm" asChild>
                          <Link to={`/games/${game.id}`}>
                            <Eye className="mr-2 h-4 w-4" />
                            Подробности
                          </Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>

              {/* Pagination */}
              <div className="flex items-center justify-between mt-4">
                <p className="text-sm text-muted-foreground">
                  Показано {games.length} из {total} игр
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setOffset(Math.max(0, offset - limit))}
                    disabled={offset === 0}
                  >
                    Назад
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setOffset(offset + limit)}
                    disabled={offset + limit >= total}
                  >
                    Вперед
                  </Button>
                </div>
              </div>
            </>
          ) : (
            <p className="text-muted-foreground text-center py-8">
              Игры пока не созданы
            </p>
          )}
        </CardContent>
      </Card>

      {/* Create Game Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Создать новую игру</DialogTitle>
            <DialogDescription>
              Заполните параметры игры
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmit}>
            <div className="grid gap-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="type">Тип игры</Label>
                  <Select
                    value={formData.type}
                    onValueChange={(value) => setFormData({ ...formData, type: value as GameType })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Выберите тип" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="cash_time">Cash (Time)</SelectItem>
                      <SelectItem value="cash_open">Cash (Open)</SelectItem>
                      <SelectItem value="tournament">Турнир</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="money_model">Деньги</Label>
                  <Select
                    value={formData.money_model}
                    onValueChange={(value) => setFormData({ ...formData, money_model: value as MoneyModel })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Выберите" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="real">Реальные деньги</SelectItem>
                      <SelectItem value="chip">Чипы</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="buy_in_amount">Buy-in amount (₽)</Label>
                  <Input
                    id="buy_in_amount"
                    type="number"
                    value={formData.buy_in_amount}
                    onChange={(e) => setFormData({ ...formData, buy_in_amount: parseFloat(e.target.value) || 0 })}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="start_time">Время начала</Label>
                  <Input
                    id="start_time"
                    type="datetime-local"
                    value={formData.start_time}
                    onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="min_players">Min players</Label>
                  <Input
                    id="min_players"
                    type="number"
                    value={formData.min_players}
                    onChange={(e) => setFormData({ ...formData, min_players: parseInt(e.target.value) || 2 })}
                    required
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="max_players">Max players</Label>
                  <Input
                    id="max_players"
                    type="number"
                    value={formData.max_players}
                    onChange={(e) => setFormData({ ...formData, max_players: parseInt(e.target.value) || 10 })}
                    required
                  />
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <input
                  id="rebuy_allowed"
                  type="checkbox"
                  checked={formData.rebuy_allowed}
                  onChange={(e) => setFormData({ ...formData, rebuy_allowed: e.target.checked })}
                  className="h-4 w-4"
                />
                <Label htmlFor="rebuy_allowed">Разрешить rebuy</Label>
              </div>
              {formData.rebuy_allowed && (
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="rebuy_amount">Rebuy amount (₽)</Label>
                    <Input
                      id="rebuy_amount"
                      type="number"
                      value={formData.rebuy_amount}
                      onChange={(e) => setFormData({ ...formData, rebuy_amount: parseFloat(e.target.value) || 0 })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="max_rebuys_per_player">Max rebuys per player</Label>
                    <Input
                      id="max_rebuys_per_player"
                      type="number"
                      value={formData.max_rebuys_per_player}
                      onChange={(e) => setFormData({ ...formData, max_rebuys_per_player: parseInt(e.target.value) || 0 })}
                    />
                  </div>
                </div>
              )}
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                Отмена
              </Button>
              <Button type="submit" disabled={createGameMutation.isPending}>
                {createGameMutation.isPending ? 'Создание...' : 'Создать'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

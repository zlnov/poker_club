import { useParams, useNavigate } from 'react-router-dom'
import { useGame, useBuyIn, useSetChips, useFinishGame } from '@/features/games/hooks'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/shared/ui/table'
import { Badge } from '@/shared/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { useAuthStore } from '@/features/auth/store'
import { GameParticipant, MoneyModel } from '@/entities/game/types'
import { format } from 'date-fns'
import { ru } from 'date-fns/locale'
import { Plus, Trophy } from 'lucide-react'
import { useState } from 'react'

const formatDate = (dateString: string) => {
  try {
    return format(new Date(dateString), 'dd MMM yyyy, HH:mm', { locale: ru })
  } catch {
    return dateString
  }
}

const getMoneyModelLabel = (model: MoneyModel) => {
  return model === 'real' ? 'Реальные деньги' : 'Чипы'
}

export function GameDetailsPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const gameId = parseInt(id || '0')
  const user = useAuthStore((state) => state.user)

  const { data, isLoading } = useGame(gameId)
  const buyInMutation = useBuyIn()
  const setChipsMutation = useSetChips()
  const finishGameMutation = useFinishGame()

  const game = data?.game
  const participants: GameParticipant[] = data?.participants || []

  const [selectedPlayerId, setSelectedPlayerId] = useState<number | null>(null)
  const [chipsAmount, setChipsAmount] = useState('')
  const [isBuyInDialogOpen, setIsBuyInDialogOpen] = useState(false)
  const [isSetChipsDialogOpen, setIsSetChipsDialogOpen] = useState(false)
  const [isFinishDialogOpen, setIsFinishDialogOpen] = useState(false)

  const isBanker = user?.id === game?.banker_id

  const handleBuyIn = () => {
    if (!selectedPlayerId) return
    buyInMutation.mutate(
      { game_id: gameId, player_id: selectedPlayerId },
      {
        onSuccess: () => {
          setIsBuyInDialogOpen(false)
          setSelectedPlayerId(null)
        },
      }
    )
  }

  const handleSetChips = () => {
    if (!selectedPlayerId || !chipsAmount) return
    setChipsMutation.mutate(
      { game_id: gameId, player_id: selectedPlayerId, chips: parseFloat(chipsAmount) },
      {
        onSuccess: () => {
          setIsSetChipsDialogOpen(false)
          setSelectedPlayerId(null)
          setChipsAmount('')
        },
      }
    )
  }

  const handleFinishGame = () => {
    finishGameMutation.mutate(
      { game_id: gameId },
      {
        onSuccess: () => {
          setIsFinishDialogOpen(false)
          navigate('/games')
        },
      }
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    )
  }

  if (!game) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Игра не найдена</p>
      </div>
    )
  }

  const totalBuyIns = participants.reduce((sum, p) => sum + p.buy_in_count * game.buy_in_amount, 0)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Игра #{game.id}</h1>
          <p className="text-muted-foreground">
            {getMoneyModelLabel(game.money_model)} • {game.type === 'tournament' ? 'Турнир' : 'Cash Game'}
          </p>
        </div>
        <div className="flex gap-2">
          {isBanker && !game.end_time && (
            <>
              <Button onClick={() => setIsBuyInDialogOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Buy-In / Rebuy
              </Button>
              <Button variant="outline" onClick={() => setIsFinishDialogOpen(true)}>
                <Trophy className="mr-2 h-4 w-4" />
                Завершить
              </Button>
            </>
          )}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Статус</CardTitle>
          </CardHeader>
          <CardContent>
            <Badge variant={game.end_time ? 'secondary' : 'success'}>
              {game.end_time ? 'Завершена' : 'Активна'}
            </Badge>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Buy-in</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{game.buy_in_amount} ₽</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Начало</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm">{formatDate(game.start_time)}</div>
          </CardContent>
        </Card>
        {game.end_time && (
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Окончание</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-sm">{formatDate(game.end_time)}</div>
            </CardContent>
          </Card>
        )}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Участники</CardTitle>
          <CardDescription>
            Всего: {participants.length} • Общий buy-in: {totalBuyIns} ₽
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <p className="text-muted-foreground text-center py-8">Загрузка...</p>
          ) : participants.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Игрок</TableHead>
                  <TableHead>Buy-in</TableHead>
                  <TableHead>Rebuy</TableHead>
                  <TableHead>Чипы</TableHead>
                  <TableHead>Место</TableHead>
                  {isBanker && !game.end_time && <TableHead>Действия</TableHead>}
                </TableRow>
              </TableHeader>
              <TableBody>
                {participants.map((participant: GameParticipant) => (
                  <TableRow key={participant.player_id}>
                    <TableCell>{participant.player_name}</TableCell>
                    <TableCell>{participant.buy_in_count}</TableCell>
                    <TableCell>{participant.rebuy_count}</TableCell>
                    <TableCell>{(participant.chips_end || 0).toFixed(2)}</TableCell>
                    <TableCell>
                      {participant.place ? (
                        <Badge variant="outline">#{participant.place}</Badge>
                      ) : (
                        '-'
                      )}
                    </TableCell>
                    {isBanker && !game.end_time && (
                      <TableCell>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => {
                            setSelectedPlayerId(participant.player_id)
                            setIsSetChipsDialogOpen(true)
                          }}
                        >
                          Установить чипы
                        </Button>
                      </TableCell>
                    )}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <p className="text-muted-foreground text-center py-8">
              Нет участников
            </p>
          )}
        </CardContent>
      </Card>

      {/* BuyIn Dialog */}
      <Dialog open={isBuyInDialogOpen} onOpenChange={setIsBuyInDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Buy-In / Rebuy</DialogTitle>
            <DialogDescription>
              Выберите игрока для добавления в игру
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label>Игрок</Label>
            <select
              className="w-full p-2 border rounded-md"
              value={selectedPlayerId || ''}
              onChange={(e) => setSelectedPlayerId(parseInt(e.target.value))}
            >
              <option value="">Выберите игрока</option>
              {participants.map((p) => (
                <option key={p.player_id} value={p.player_id}>
                  {p.player_name}
                </option>
              ))}
            </select>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsBuyInDialogOpen(false)}>
              Отмена
            </Button>
            <Button onClick={handleBuyIn} disabled={!selectedPlayerId || buyInMutation.isPending}>
              {buyInMutation.isPending ? 'Выполнение...' : 'Подтвердить'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Set Chips Dialog */}
      <Dialog open={isSetChipsDialogOpen} onOpenChange={setIsSetChipsDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Установить количество чипов</DialogTitle>
            <DialogDescription>
              Введите финальное количество чипов игрока
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Label>Количество чипов</Label>
            <Input
              type="number"
              step="0.01"
              value={chipsAmount}
              onChange={(e) => setChipsAmount(e.target.value)}
              placeholder="0.00"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsSetChipsDialogOpen(false)}>
              Отмена
            </Button>
            <Button onClick={handleSetChips} disabled={!chipsAmount || setChipsMutation.isPending}>
              {setChipsMutation.isPending ? 'Сохранение...' : 'Сохранить'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Finish Game Dialog */}
      <Dialog open={isFinishDialogOpen} onOpenChange={setIsFinishDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Завершить игру</DialogTitle>
            <DialogDescription>
              Вы уверены, что хотите завершить эту игру? Это действие нельзя отменить.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsFinishDialogOpen(false)}>
              Отмена
            </Button>
            <Button variant="destructive" onClick={handleFinishGame} disabled={finishGameMutation.isPending}>
              {finishGameMutation.isPending ? 'Завершение...' : 'Завершить игру'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

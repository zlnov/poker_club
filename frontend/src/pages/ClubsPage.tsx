import { useState } from 'react'
import { useClubs, useClubMembers, useCreateClub, useApproveMember, useRejectMember } from '@/features/clubs/hooks'
import { Button } from '@/shared/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/shared/ui/table'
import { Badge } from '@/shared/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Users, Plus, CheckCircle, XCircle } from 'lucide-react'

export function ClubsPage() {
  const { data: clubs, isLoading } = useClubs()
  const createClubMutation = useCreateClub()
  const approveMemberMutation = useApproveMember()
  const rejectMemberMutation = useRejectMember()

  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [selectedClubId, setSelectedClubId] = useState<number | null>(null)
  const [clubName, setClubName] = useState('')

  const { data: members } = useClubMembers(selectedClubId || 0)

  const handleCreateClub = (e: React.FormEvent) => {
    e.preventDefault()
    createClubMutation.mutate({ name: clubName }, {
      onSuccess: () => {
        setIsCreateDialogOpen(false)
        setClubName('')
      }
    })
  }

  const handleApproveMember = (memberId: number) => {
    if (selectedClubId) {
      approveMemberMutation.mutate({ clubId: selectedClubId, memberId })
    }
  }

  const handleRejectMember = (memberId: number) => {
    if (selectedClubId) {
      rejectMemberMutation.mutate({ clubId: selectedClubId, memberId })
    }
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
        <h1 className="text-3xl font-bold">Клубы</h1>
        <Button onClick={() => setIsCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Создать клуб
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {clubs?.map((club) => (
          <Card key={club.id} className="cursor-pointer hover:shadow-md transition-shadow"
            onClick={() => setSelectedClubId(club.id)}
          >
            <CardHeader>
              <CardTitle>{club.name}</CardTitle>
              <CardDescription>
                Создан: {new Date(club.created_at).toLocaleDateString('ru-RU')}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Users className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">
                    {members?.length || 0} участников
                  </span>
                </div>
                <Button variant="outline" size="sm">
                  Управлять
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {clubs && clubs.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-8 text-muted-foreground">
            <Users className="h-12 w-12 mb-4" />
            <p>Клубы пока не созданы</p>
          </CardContent>
        </Card>
      )}

      {/* Create Club Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Создать новый клуб</DialogTitle>
            <DialogDescription>
              Введите название клуба
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateClub}>
            <div className="py-4">
              <Label htmlFor="clubName">Название клуба</Label>
              <Input
                id="clubName"
                value={clubName}
                onChange={(e) => setClubName(e.target.value)}
                placeholder="Название клуба"
                required
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setIsCreateDialogOpen(false)}>
                Отмена
              </Button>
              <Button type="submit" disabled={createClubMutation.isPending}>
                {createClubMutation.isPending ? 'Создание...' : 'Создать'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Members Management Dialog */}
      <Dialog open={!!selectedClubId} onOpenChange={(open) => !open && setSelectedClubId(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Управление участниками</DialogTitle>
            <DialogDescription>
              {clubs?.find(c => c.id === selectedClubId)?.name}
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            {members && members.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Игрок</TableHead>
                    <TableHead>Роль</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Действия</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {members.map((member) => (
                    <TableRow key={member.id}>
                      <TableCell>
                        <div>
                          <p className="font-medium">{member.player.first_name} {member.player.last_name}</p>
                          <p className="text-sm text-muted-foreground">@{member.player.nickname}</p>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant={member.role === 'admin' ? 'default' : 'secondary'}>
                          {member.role === 'admin' ? 'Админ' : 'Участник'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant={
                          member.status === 'active' ? 'success' :
                          member.status === 'pending' ? 'warning' : 'destructive'
                        }>
                          {member.status === 'active' ? 'Активен' :
                           member.status === 'pending' ? 'Ожидает' : 'Заблокирован'}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {member.status === 'pending' && (
                          <div className="flex space-x-2">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => handleApproveMember(member.player_id)}
                              disabled={approveMemberMutation.isPending}
                            >
                              <CheckCircle className="h-4 w-4 mr-1" />
                              Одобрить
                            </Button>
                            <Button
                              size="sm"
                              variant="destructive"
                              onClick={() => handleRejectMember(member.player_id)}
                              disabled={rejectMemberMutation.isPending}
                            >
                              <XCircle className="h-4 w-4 mr-1" />
                              Отклонить
                            </Button>
                          </div>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : (
              <p className="text-muted-foreground text-center py-8">
                Нет участников
              </p>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

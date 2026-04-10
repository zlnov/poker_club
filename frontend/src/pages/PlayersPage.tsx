import { useClubMembers } from '@/features/clubs/hooks'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/shared/ui/table'
import { Badge } from '@/shared/ui/badge'

export function PlayersPage() {
  const currentClubId = 1 // TODO: получить из глобального состояния
  const { data: members, isLoading } = useClubMembers(currentClubId)

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
        <h1 className="text-3xl font-bold">Игроки</h1>
        <p className="text-muted-foreground">
          Всего участников: {members?.length || 0}
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Список участников клуба</CardTitle>
          <CardDescription>Просмотр всех участников клуба и их ролей</CardDescription>
        </CardHeader>
        <CardContent>
          {members && members.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Игрок</TableHead>
                  <TableHead>Роль</TableHead>
                  <TableHead>Статус</TableHead>
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
                        member.status === 'active' ? 'default' :
                        member.status === 'pending' ? 'outline' : 'destructive'
                      }>
                        {member.status === 'active' ? 'Активен' :
                         member.status === 'pending' ? 'Ожидание' : 'Заблокирован'}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              В клубе пока нет участников
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

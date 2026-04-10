import { useAuthStore } from '@/features/auth/store'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Button } from '@/shared/ui/button'
import { Shield, Users, Settings as SettingsIcon } from 'lucide-react'

export function SettingsPage() {
  const user = useAuthStore((state) => state.user)
  const isAdmin = user?.role === 'admin'

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Настройки</h1>
      </div>

      {!isAdmin ? (
        <Card>
          <CardHeader>
            <CardTitle>Доступ ограничен</CardTitle>
            <CardDescription>
              Настройки клуба доступны только администраторам
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-center py-8 text-muted-foreground">
              <Shield className="h-12 w-12 mb-4" />
              <p>У вас нет прав для просмотра этой страницы</p>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Управление клубом</CardTitle>
              <CardDescription>Настройки клуба и участников</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center space-x-3">
                    <Users className="h-8 w-8 text-muted-foreground" />
                    <div>
                      <p className="font-medium">Управление участниками</p>
                      <p className="text-sm text-muted-foreground">
                        Одобрение/отклонение заявок, управление ролями
                      </p>
                    </div>
                  </div>
                  <Button variant="outline">Открыть</Button>
                </div>
                <div className="flex items-center justify-between p-4 border rounded-lg">
                  <div className="flex items-center space-x-3">
                    <SettingsIcon className="h-8 w-8 text-muted-foreground" />
                    <div>
                      <p className="font-medium">Параметры игр</p>
                      <p className="text-sm text-muted-foreground">
                        Настройки buy-in, rebuy, лимитов
                      </p>
                    </div>
                  </div>
                  <Button variant="outline">Открыть</Button>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Информация о системе</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Версия frontend:</span>
                  <span>1.0.0 MVP</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Роль:</span>
                  <span className="capitalize">{user?.role}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Пользователь:</span>
                  <span>{user?.nickname}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

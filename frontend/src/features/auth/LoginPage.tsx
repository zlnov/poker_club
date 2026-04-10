import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/shared/ui/card'
import { login } from './api'
import { useAuthStore } from './store'
import { LoginPayload } from '@/entities/auth/types'

export function LoginPage() {
  const navigate = useNavigate()
  const loginStore = useAuthStore()
  const queryClient = useQueryClient()

  const [phone_number, setPhone] = useState('')
  const [password, setPassword] = useState('')

  const mutation = useMutation({
    mutationFn: (payload: LoginPayload) => login(payload),
    onSuccess: ({ user, token, refresh_token }) => {
      loginStore.login(user, token, refresh_token)
      queryClient.clear()
      navigate('/dashboard')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    mutation.mutate({ phone_number, password })
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">Poker Club</CardTitle>
          <CardDescription className="text-center">
            Введите ваш телефон и пароль для входа
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="phone_number">Телефон</Label>
              <Input
                id="phone_number"
                type="tel"
                placeholder="+7 (999) 123-45-67"
                value={phone_number}
                onChange={(e) => setPhone(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Пароль</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            {mutation.isError && (
              <p className="text-sm text-destructive">
                {(mutation.error as any)?.response?.data?.error || 'Ошибка входа'}
              </p>
            )}
          </CardContent>
          <CardFooter>
            <Button type="submit" className="w-full" disabled={mutation.isPending}>
              {mutation.isPending ? 'Вход...' : 'Войти'}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}

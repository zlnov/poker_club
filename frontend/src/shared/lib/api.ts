import axios from 'axios'
import { getAccessToken, setAccessToken, clearAuth } from './auth'
import { authStoreInstance } from '@/features/auth/store'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

  api.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config as any

    // При 401 попробуем обновить токен, если есть refresh_token
    if (error.response?.status === 401 && !originalRequest._retry) {
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          originalRequest._retry = true
           // импортируем функцию refreshToken из модуля API аутентификации
          const { access, refresh } = await import('@/features/auth/api').then(m => m.refreshToken())
          // обновляем access token в хранилище
          setAccessToken(access)
          if (refresh) {
            localStorage.setItem('refresh_token', refresh)
          }
          originalRequest.headers.Authorization = `Bearer ${access}`
          return api(originalRequest)
        } catch {
          authStoreInstance.getState().logout()
          return Promise.reject(new Error('Unauthorized'))
        }
      }
      authStoreInstance.getState().logout()
      return Promise.reject(new Error('Unauthorized'))
    }

    return Promise.reject(error)
  }
)

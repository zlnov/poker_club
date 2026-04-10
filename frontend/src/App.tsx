import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './features/auth/store'
import { ProtectedRoute } from './shared/components/ProtectedRoute'
import { Layout } from './app/Layout'
import { LoginPage } from './features/auth/LoginPage'
import { DashboardPage } from './pages/DashboardPage'
import { GamesPage } from './pages/GamesPage'
import { GameDetailsPage } from './pages/GameDetailsPage'
// import { PlayersPage } from './pages/PlayersPage' // Временно скрыт - отсутствует GET /players
import { LeaderboardPage } from './pages/LeaderboardPage'
import { ClubsPage } from './pages/ClubsPage'
import { SettingsPage } from './pages/SettingsPage'

function App() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)

  return (
    <Routes>
      <Route
        path="/login"
        element={isAuthenticated ? <Navigate to="/dashboard" /> : <LoginPage />}
      />
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/games" element={<GamesPage />} />
        <Route path="/games/:id" element={<GameDetailsPage />} />
        {/* <Route path="/players" element={<PlayersPage />} /> // Временно скрыт - отсутствует GET /players */}
        <Route path="/leaderboard" element={<LeaderboardPage />} />
        <Route path="/clubs" element={<ClubsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Route>
      <Route path="/" element={<Navigate to="/dashboard" />} />
      <Route path="*" element={<Navigate to="/dashboard" />} />
    </Routes>
  )
}

export default App

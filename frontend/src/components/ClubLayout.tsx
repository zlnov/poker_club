/**
 * Club layout component.
 *
 * Provides tab-based navigation for club-specific routes.
 * Uses Mantine Tabs with each tab linked to its own route.
 *
 * See 05_FE_UX.md section 5 (Routing Structure) and
 * section 6 (Club Navigation).
 */

import { Tabs, TabsList, TabsTab, TabsPanel } from '@mantine/core'
import {
  IconHome,
  IconChessKing,
  IconUsers,
  IconChartBar,
  IconSettings,
} from '@tabler/icons-react'
import { useNavigate, useLocation, useParams, Outlet } from 'react-router-dom'

/**
 * Club layout component.
 *
 * Renders club navigation tabs and an outlet for club sub-routes.
 * The active tab is determined by the current route.
 */
export function ClubLayout() {
  const location = useLocation()
  const params = useParams()
  const navigate = useNavigate()
  const clubId = params.clubId

  // Determine the active tab based on the current path
  const getActiveTab = (): string => {
    const path = location.pathname
    if (path.includes('/members')) return 'members'
    if (path.includes('/games')) return 'games'
    if (path.includes('/statistics')) return 'statistics'
    if (path.includes('/settings')) return 'settings'
    return 'dashboard'
  }

  const activeTab = getActiveTab()
  const basePath = `/clubs/${clubId}`

  const handleTabChange = (tab: string | null) => {
    if (!tab) return
    const tabPath = tab === 'dashboard' ? basePath : `${basePath}/${tab}`
    navigate(tabPath)
  }

  return (
    <Tabs
      value={activeTab}
      variant="pills"
      radius="md"
      onChange={handleTabChange}
    >
      <TabsList grow>
        <TabsTab value="dashboard" leftSection={<IconHome size={16} />}>
          Dashboard
        </TabsTab>
        <TabsTab value="games" leftSection={<IconChessKing size={16} />}>
          Games
        </TabsTab>
        <TabsTab value="members" leftSection={<IconUsers size={16} />}>
          Members
        </TabsTab>
        <TabsTab value="statistics" leftSection={<IconChartBar size={16} />}>
          Statistics
        </TabsTab>
        <TabsTab value="settings" leftSection={<IconSettings size={16} />}>
          Settings
        </TabsTab>
      </TabsList>

      <TabsPanel value="dashboard" pt="md">
        <Outlet />
      </TabsPanel>
      <TabsPanel value="games" pt="md">
        <Outlet />
      </TabsPanel>
      <TabsPanel value="members" pt="md">
        <Outlet />
      </TabsPanel>
      <TabsPanel value="statistics" pt="md">
        <Outlet />
      </TabsPanel>
      <TabsPanel value="settings" pt="md">
        <Outlet />
      </TabsPanel>
    </Tabs>
  )
}

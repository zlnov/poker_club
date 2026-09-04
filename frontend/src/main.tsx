/**
 * Application entry point.
 *
 * Initializes the React application with:
 * - AppProviders (MantineProvider + QueryClientProvider + Notifications)
 * - RouterProvider (React Router for client-side routing)
 *
 * See 04_FE_SPEC.md section 9 (Frontend Architecture).
 */

import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { AppProviders } from './app/providers'
import { router } from './app/router'
import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'
import './styles/index.css'

createRoot(document.getElementById('root')!).render(
  <AppProviders>
    <RouterProvider router={router} />
  </AppProviders>,
)

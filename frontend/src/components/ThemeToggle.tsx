/**
 * ThemeToggle — a styled toggle button for switching between light and dark mode.
 *
 * Uses Mantine's useMantineColorScheme to toggle the color scheme.
 * Renders as an ActionIcon with a sun/moon icon.
 *
 * See 05_FE_UX.md section 18 (Telegram Theme) and
 * 04_FE_SPEC.md section 22 (Telegram Theme).
 */

import { ActionIcon, useMantineColorScheme } from '@mantine/core'
import { IconSun, IconMoonStars } from '@tabler/icons-react'
import { forwardRef } from 'react'

export interface ThemeToggleProps {
  /** Size of the toggle button. */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

/**
 * Theme toggle button that switches between light and dark mode.
 *
 * Usage:
 *   <ThemeToggle />
 */
export const ThemeToggle = forwardRef<HTMLButtonElement, ThemeToggleProps>(
  ({ size = 'md' }, ref) => {
    const { colorScheme, toggleColorScheme } = useMantineColorScheme()

    const isDark = colorScheme === 'dark'

    return (
      <ActionIcon
        ref={ref}
        variant="subtle"
        size={size}
        radius="md"
        onClick={() => toggleColorScheme()}
        aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
        title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
      >
        {isDark ? (
          <IconSun size={18} stroke={1.5} />
        ) : (
          <IconMoonStars size={18} stroke={1.5} />
        )}
      </ActionIcon>
    )
  },
)

ThemeToggle.displayName = 'ThemeToggle'

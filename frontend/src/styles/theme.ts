/**
 * Mantine theme configuration for the Poker Club Frontend.
 *
 * Defines the Poker Club Design System colors, typography, spacing,
 * and border radius based on the approved color palette.
 *
 * See 05_FE_UX.md section 2 (UI Library / Design System) and
 * 04_FE_SPEC.md section 8 (Technology Stack — Mantine approved).
 */

import type { MantineThemeOverride } from '@mantine/core'

/**
 * Poker Club color palette.
 *
 * Each color is a MantineColorsTuple (readonly tuple of 10 strings).
 * Colors are centralized here so individual components reference theme
 * colors rather than hardcoding values.
 */
export const pokerClubColors = {
  // Violet / accent palette — primary brand color
  violet: [
    '#f5f3ff',
    '#ede9fe',
    '#ddd6fe',
    '#c4b5fd',
    '#c084fc',
    '#a855f7',
    '#8b5cf6',
    '#7c3aed',
    '#6d28d9',
    '#5b21b6',
  ],
  // Success
  success: [
    '#dcfce8',
    '#bbf7d0',
    '#86efac',
    '#4ade80',
    '#22c55e',
    '#10b981',
    '#16a34a',
    '#15803d',
    '#166534',
    '#14532d',
  ],
  // Error
  error: [
    '#fee2e2',
    '#fecaca',
    '#fca5a5',
    '#f87171',
    '#ef4444',
    '#ef4444',
    '#dc2626',
    '#b91c1c',
    '#991b1b',
    '#7f1d1d',
  ],
  // Warning
  warning: [
    '#fef3c7',
    '#fde68a',
    '#fcd34d',
    '#fbbf24',
    '#f59e0b',
    '#f59e0b',
    '#d97706',
    '#b45309',
    '#92400e',
    '#78350f',
  ],
  // Info
  info: [
    '#dbeafe',
    '#bfdbfe',
    '#93c5fd',
    '#60a5fa',
    '#3b82f6',
    '#3b82f6',
    '#2563eb',
    '#1d4ed8',
    '#1e40af',
    '#1e3a8a',
  ],
  // Gray / neutral palette
  gray: [
    '#f9fafb',
    '#f3f4f6',
    '#e5e7eb',
    '#d1d5db',
    '#9ca3af',
    '#6b7280',
    '#4b5563',
    '#374151',
    '#1f2937',
    '#111827',
  ],
} as const satisfies Record<
  string,
  readonly [
    string,
    string,
    string,
    string,
    string,
    string,
    string,
    string,
    string,
    string,
  ]
>

/**
 * Mantine theme override for the Poker Club Frontend.
 *
 * Configures:
 * - Primary color (violet/accent)
 * - Color scheme support (light/dark)
 * - Typography (font family, heading sizes)
 * - Spacing and border radius
 * - Component variants for consistency
 * - Shadows for elevation
 *
 * See 05_FE_UX.md section 2 (Design System).
 */
export const pokerClubTheme: MantineThemeOverride = {
  colors: pokerClubColors,
  primaryColor: 'violet',
  fontFamily:
    "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
  fontFamilyMonospace:
    "ui-monospace, 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', monospace",
  headings: {
    fontFamily:
      "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
    fontWeight: '500',
    sizes: {
      h1: { fontSize: '2.25rem', lineHeight: '2.5rem' },
      h2: { fontSize: '1.875rem', lineHeight: '2.25rem' },
      h3: { fontSize: '1.5rem', lineHeight: '2rem' },
      h4: { fontSize: '1.25rem', lineHeight: '1.75rem' },
      h5: { fontSize: '1.125rem', lineHeight: '1.5rem' },
      h6: { fontSize: '1rem', lineHeight: '1.5rem' },
    },
  },
  radius: {
    xs: '0.25rem',
    sm: '0.375rem',
    md: '0.5rem',
    lg: '0.75rem',
    xl: '1rem',
  },
  spacing: {
    xs: '0.25rem',
    sm: '0.5rem',
    md: '1rem',
    lg: '1.5rem',
    xl: '2rem',
  },
  shadows: {
    xs: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 2px -1px rgba(0, 0, 0, 0.03)',
    sm: '0 1px 2px 0 rgba(0, 0, 0, 0.05), 0 1px 2px -1px rgba(0, 0, 0, 0.05)',
    md: '0 1px 4px 0 rgba(0, 0, 0, 0.06), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 4px 6px -1px rgba(0, 0, 0, 0.07), 0 2px 4px -2px rgba(0, 0, 0, 0.07)',
    xl: '0 10px 15px -3px rgba(0, 0, 0, 0.08), 0 4px 6px -4px rgba(0, 0, 0, 0.08)',
  },
  components: {
    Button: {
      styles: {
        root: {
          fontWeight: 500,
        },
      },
      defaultProps: {
        radius: 'md',
      },
    },
    ActionIcon: {
      defaultProps: {
        radius: 'md',
        variant: 'subtle',
      },
    },
    NavLink: {
      styles: {
        body: {
          fontWeight: 500,
        },
      },
    },
    Card: {
      defaultProps: {
        radius: 'md',
        withBorder: true,
      },
    },
    Paper: {
      defaultProps: {
        radius: 'md',
        withBorder: true,
      },
    },
    Input: {
      defaultProps: {
        radius: 'md',
      },
    },
    TextInput: {
      defaultProps: {
        radius: 'md',
      },
    },
    PasswordInput: {
      defaultProps: {
        radius: 'md',
      },
    },
    Select: {
      defaultProps: {
        radius: 'md',
      },
    },
    Textarea: {
      defaultProps: {
        radius: 'md',
      },
    },
    Container: {
      defaultProps: {
        size: 'lg',
      },
    },
    Title: {
      defaultProps: {
        fw: 500,
      },
    },
  },
}

export default pokerClubTheme

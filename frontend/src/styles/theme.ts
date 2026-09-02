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
 * Based on the approved design tokens from src/styles/index.css.
 * The primary color is violet, matching the --accent CSS variable.
 *
 * Each color is a MantineColorsTuple (readonly tuple of 10 strings).
 */
export const pokerClubColors = {
  // Violet / accent palette (matches --accent)
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
  // Success (matches --success)
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
  // Error (matches --error)
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
  // Warning (matches --warning)
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
  // Info (matches --info)
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
  // Gray / neutral palette (matches --border, --muted, etc.)
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
 *
 * See 05_FE_UX.md section 2 (Design System).
 */
export const pokerClubTheme: MantineThemeOverride = {
  colors: pokerClubColors,
  primaryColor: 'violet',
  // Use the CSS font stack from index.css
  fontFamily:
    "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
  fontFamilyMonospace:
    "ui-monospace, 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'Liberation Mono', 'Courier New', monospace",
  headings: {
    fontFamily:
      "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
    fontWeight: '500',
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
}

export default pokerClubTheme

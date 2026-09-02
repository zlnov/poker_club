/**
 * Base styles for the Poker Club Frontend.
 *
 * Design System foundation (see 05_FE_UX.md section 20–60).
 * Uses CSS variables for theming with light/dark mode support.
 *
 * See 04_FE_SPEC.md section 21 (responsive design), section 691–709
 * (Telegram theme), section 890–902 (browser compatibility).
 */

/**
 * CSS Variables — core design tokens.
 * Themable via the `data-theme` attribute and Telegram theme parameters.
 */
export const cssVariables = {
  /** Background color. */
  bg: '#ffffff',
  /** Text color. */
  text: '#1f2937',
  /** Border color. */
  border: '#d1d5db',
  /** Accent color (primary action). */
  accent: '#a855f7',
  /** Accent background (subtle). */
  accentBg: 'rgba(168, 85, 247, 0.1)',
  /** Subtle border. */
  subtleBorder: '#e5e7eb',
  /** Muted text. */
  muted: '#6b7280',
  /** Success color. */
  success: '#10b981',
  /** Error color. */
  error: '#ef4444',
  /** Warning color. */
  warning: '#f59e0b',
  /** Info color. */
  info: '#3b82f6',
}

/** Responsive breakpoints based on 05_FE_UX.md section 431–451. */
export const breakpoints = {
  /** Mobile (primary priority, 05_FE_UX.md section 436). */
  mobile: '640px',
  /** Tablet. */
  tablet: '768px',
  /** Desktop. */
  desktop: '1024px',
}

/** CSS media queries for responsive design. */
export const media = {
  maxMobile: `(max-width: ${breakpoints.mobile})`,
  maxTablet: `(max-width: ${breakpoints.tablet})`,
  maxDesktop: `(max-width: ${breakpoints.desktop})`,
  minMobile: `(min-width: ${breakpoints.mobile})`,
  minTablet: `(min-width: ${breakpoints.tablet})`,
}

/** CSS animation utilities. */
export const animation = {
  fast: 'all 0.15s ease-in-out',
  medium: 'all 0.25s ease-in-out',
  slow: 'all 0.4s ease-in-out',
}

/** z-index stack for layering. */
export const zIndex = {
  base: 0,
  modal: 1000,
  dropdown: 1100,
  popover: 1200,
  tooltip: 1300,
  fixed: 2000,
}

/**
 * Applies Telegram theme parameters to the document CSS variables.
 *
 * When running in Telegram Mini App mode, the theme colors from
 * `window.Telegram.WebApp.theme` are applied.
 * In Standard Web mode, the default theme is used.
 *
 * See 04_FE_SPEC.md section 703 (Standard Web fallback theme).
 * See 05_FE_UX.md section 412–429 (Telegram theme).
 */
export function applyTelegramThemeVariables(theme?: {
  bgColor: string
  textColor: string
}): void {
  const html = document.documentElement

  if (theme) {
    html.style.setProperty('--bg', theme.bgColor)
    html.style.setProperty('--text', theme.textColor)
  } else {
    // Standard Web — use default theme
    html.style.setProperty('--bg', cssVariables.bg)
    html.style.setProperty('--text', cssVariables.text)
  }
}

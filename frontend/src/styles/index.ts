/**
 * Base styles for the Poker Club Frontend.
 *
 * Design System foundation (see 05_FE_UX.md section 2).
 * Uses CSS variables for theming with light/dark mode support.
 *
 * See 04_FE_SPEC.md section 21 (responsive design), section 22 (Telegram theme),
 * section 34 (accessibility), section 890–902 (browser compatibility).
 */

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
 * See 04_FE_SPEC.md section 22 (Telegram Theme) and
 * 05_FE_UX.md section 18 (Telegram Theme).
 */
export function applyTelegramThemeVariables(theme?: {
  bgColor: string
  textColor: string
}): void {
  const html = document.documentElement

  if (theme) {
    html.style.setProperty('--mantine-color-body', theme.bgColor)
    html.style.setProperty('--mantine-color-text', theme.textColor)
  }
  // Standard Web mode — Mantine handles light/dark via colorScheme
}

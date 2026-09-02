/**
 * TelegramMode component.
 *
 * A Telegram-safe component that conditionally renders Telegram-specific
 * UI vs Standard Web UI based on the detected environment.
 *
 * See 04_FE_SPEC.md section 7 (Telegram and Standard Web compatibility).
 *
 * Usage:
 *   <TelegramMode
 *     env={telegramEnv}
 *     telegram={({ webApp }) => <TelegramOnlyUI ... />}
 *     standard={({ isStandardWeb }) => <StandardWebUI ... />}
 *   />
 */

import type { ReactNode } from 'react'
import type { TelegramEnvironmentInfo, TelegramWebAppAPI } from './index'

export interface TelegramModeProps {
  /** Telegram environment info. */
  env: TelegramEnvironmentInfo
  /** Render prop for Telegram Mini App mode. */
  telegram: (props: { webApp?: TelegramWebAppAPI }) => ReactNode
  /** Render prop for Standard Web mode. */
  standard: (props: { isStandardWeb: boolean }) => ReactNode
}

/**
 * Renders the appropriate UI based on the Telegram environment.
 *
 * The component detects the environment and renders either the Telegram
 * specific UI or the Standard Web fallback UI.
 *
 * The Telegram-specific code is isolated in the `telegram` render prop,
 * and the Standard Web code is in the `standard` render prop.
 */
export function TelegramMode({ env, telegram, standard }: TelegramModeProps) {
  if (env.isTelegram) {
    return <>{telegram({ webApp: env.webApp })}</>
  }
  return <>{standard({ isStandardWeb: env.isStandardWeb })}</>
}

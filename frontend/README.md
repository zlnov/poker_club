# Poker Club Frontend

React + TypeScript + Vite frontend for the Poker Club management system.

## Technology Stack

- **React** — UI library
- **TypeScript** — type safety
- **Vite** — build tool
- **React Router** — client-side routing
- **TanStack Query** — server state management

## Project Structure

```
frontend/
├── src/
│   ├── app/          # Application shell (router, providers)
│   ├── pages/        # Page components
│   ├── features/     # Feature modules (added in later phases)
│   ├── components/   # Reusable UI components
│   ├── api/          # API client and error handling
│   ├── auth/         # Authentication layer (added in RM_FE_02)
│   ├── telegram/     # Telegram WebApp integration layer
│   ├── hooks/        # Custom React hooks
│   ├── types/        # Domain types
│   ├── utils/        # Utility functions
│   ├── styles/       # CSS and theme variables
│   └── vite-env.d.ts # Vite environment type declarations
├── public/           # Static assets
├── .env.development  # Development environment config
├── .env.production   # Production environment config
├── package.json
├── tsconfig.json
├── tsconfig.app.json
├── tsconfig.node.json
├── vite.config.ts
└── eslint.config.js
```

## Environment Configuration

Frontend environment configuration does **not** contain secrets.

| Variable            | Description                                      |
| ------------------- | ------------------------------------------------ |
| `VITE_API_BASE_URL` | Backend HTTP API base URL (with `/api/v1`)       |
| `VITE_ENV_NAME`     | Environment name (`development` or `production`) |

## Scripts

```bash
npm run dev      # Start development server
npm run build    # Build for production
npm run preview  # Preview production build
npm run lint     # Run linter
npm run format   # Format code with Prettier
```

## Telegram Integration

The frontend supports two modes:

1. **Standard Web** — opens in a regular browser. No Telegram API required.
2. **Telegram Mini App** — opens inside Telegram. Uses `window.Telegram.WebApp`.

The Telegram integration layer safely detects the environment and provides
fallbacks for all Telegram-specific capabilities. The app never crashes
when the Telegram WebApp API is unavailable.

## Development

```bash
npm install
npm run dev
```

The development server runs at `http://localhost:3000`.

# conclave Web UI

## Modern React Frontend

The conclave web interface has been completely rewritten with a modern React stack for a vastly improved user experience.

## What Changed

### Before (Old UI)
- Server-side rendered Go templates
- HTMX for partial page updates
- 5-second polling for debate updates
- Unoptimized Tailwind CSS loaded from CDN (~3MB)
- Clunky page refreshes
- Poor real-time experience

### After (New UI)
- **React 19** with TypeScript for type safety
- **Vite** for lightning-fast dev experience and optimized builds
- **React Router** for smooth client-side navigation
- **TanStack Query** for intelligent data fetching and caching
- **Server-Sent Events (SSE)** for real-time debate streaming
- **Optimized Tailwind CSS** with proper purging (~23KB production build)
- Smooth animations and transitions
- Real-time character-by-character streaming of AI responses

## Key Features

### 1. Real-Time Streaming
Watch AI agents debate in real-time as responses stream in character-by-character, similar to ChatGPT.

### 2. Modern SPA Experience
- Instant page navigation with no full-page reloads
- Smooth transitions and animations
- Optimistic UI updates

### 3. Performance
- **Production bundle**: ~286KB (88KB gzipped)
- **CSS**: 22KB (4.6KB gzipped)
- Fast initial load, instant subsequent navigation

### 4. Developer Experience
- **Hot Module Replacement (HMR)** during development
- **TypeScript** for type safety
- **ESLint** for code quality
- Clean component architecture

## Running the App

### Development Mode

Run frontend and backend separately for hot reload:

```bash
# Terminal 1: Frontend dev server (with proxy to backend)
cd web/app
npm run dev
# Frontend runs on http://localhost:5173 with API proxy to :8080

# Terminal 2: Backend server
make serve
# Backend runs on http://localhost:8080
```

The Vite dev server proxies API requests to the backend automatically.

### Production Mode

Build everything and run:

```bash
# Build frontend and backend together
make build

# Run the server
./bin/conclave serve
# Visit http://localhost:8080
```

The frontend is embedded in the Go binary, so you only need one executable.

## Architecture

```
/web
├── /app                    # React application
│   ├── /src
│   │   ├── /components    # Reusable UI components
│   │   ├── /pages         # Page components (Home, Debate, History)
│   │   ├── /hooks         # Custom React hooks (useDebateStream)
│   │   ├── /lib           # API client
│   │   ├── /types         # TypeScript interfaces
│   │   └── main.tsx       # App entry point
│   ├── /dist              # Production build output (generated)
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
└── /handlers               # Go HTTP handlers
    ├── handlers.go        # Main handlers
    ├── streaming.go       # SSE streaming endpoint
    └── spa.go             # Serves React SPA
```

## API Endpoints

### REST API
- `GET /api/providers` - List available AI providers
- `GET /api/debates` - List debates
- `GET /api/debates/:id` - Get debate details
- `POST /debates` - Create new debate
- `DELETE /debates/:id` - Delete debate

### Streaming
- `GET /api/debates/:id/stream` - Server-Sent Events for real-time updates

## Technology Stack

### Frontend
- **React 19** - UI framework
- **TypeScript** - Type safety
- **Vite 7** - Build tool and dev server
- **React Router 7** - Client-side routing
- **TanStack Query 5** - Server state management
- **Tailwind CSS 3** - Utility-first CSS

### Backend
- **Go 1.21+** - Backend language
- **net/http** - HTTP server
- **embed** - Embedding React build in binary

## Benefits Over Old UI

1. **Real-time streaming**: See AI responses as they're generated
2. **Better performance**: Smaller bundle, faster load times
3. **Modern UX**: Smooth animations, instant feedback
4. **Type safety**: TypeScript catches errors before runtime
5. **Better maintainability**: Component-based architecture
6. **SEO-friendly**: Can add server-side rendering if needed
7. **Mobile-friendly**: Responsive design with Tailwind
8. **Developer experience**: Hot reload, better tooling

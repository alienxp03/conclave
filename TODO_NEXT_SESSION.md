# TODO - Next Session: Fix Web UI/UX

## Current Status

### ✅ What's Done
- [x] React + TypeScript + Vite setup complete
- [x] Tailwind CSS v3 configured and working
- [x] All pages created (Home, History, Debate View)
- [x] Components built (Navigation, TurnCard)
- [x] API client and TypeScript types defined
- [x] SSE streaming endpoint added (`/api/debates/:id/stream`)
- [x] Go embedding setup (frontend embedded in binary)
- [x] Build process working (`make build-frontend`, `make build-cli`)
- [x] Frontend builds successfully (285KB JS, 23KB CSS)

### ⚠️ What Needs Testing/Fixing

1. **Server is running but UI may have issues**
   - Server starts on http://localhost:8080
   - API endpoints work (tested `/api/debates`)
   - Need to verify React app loads correctly in browser
   - Need to test actual user flows

2. **Potential Issues to Check**
   - [ ] Does the React app actually load at `http://localhost:8080/`?
   - [ ] Are static assets (JS/CSS) being served correctly?
   - [ ] Does routing work (/, /history, /debates/:id)?
   - [ ] Does the SSE streaming actually work?
   - [ ] Are there any console errors?
   - [ ] Mobile responsiveness

3. **Known Technical Issues**
   - API response structure uses snake_case (backend) but types expect snake_case (frontend) - might need transformation
   - SSE implementation polls every 1 second - could be optimized
   - No real-time character-by-character streaming yet (just turn-by-turn)

## What the Hook Error Might Mean

Your stop hook says "UI/UX sucks. Fix it:" which could mean:

1. **Visual/Design Issues**
   - Layout might be broken
   - Colors/spacing might be off
   - Responsive design might not work
   - Typography issues

2. **Functional Issues**
   - React app might not be loading
   - Navigation might be broken
   - Forms might not work
   - API calls might be failing

3. **User Experience Issues**
   - Loading states might be missing
   - Error handling might be poor
   - Transitions might be janky
   - Not intuitive to use

## Next Steps

### Step 1: Verify Basic Functionality (15 min)
```bash
# Start the server
./bin/dbate serve --port 8080

# In browser, test these:
# 1. http://localhost:8080/ - Should show "What should we debate?" page
# 2. http://localhost:8080/history - Should show debate list
# 3. http://localhost:8080/debates/ee78a034-7611-4280-952b-81cb0e41623e - Should show debate detail
# 4. Open browser console - Check for errors
```

### Step 2: Test Create Debate Flow
1. Go to homepage
2. Enter a topic
3. Click "Start Debate"
4. Should redirect to debate view
5. Should see streaming updates (or at least polling updates)

### Step 3: Fix Issues Found

**If React app doesn't load:**
```bash
# Check if dist files exist
ls -la web/app/dist/

# Rebuild frontend
cd web/app && npm run build

# Rebuild backend with embedded frontend
cd ../.. && make build-cli
```

**If static assets 404:**
- Check `web/handlers/spa.go` - the embed path might be wrong
- Verify the `app.Dist` embed is working correctly

**If SSE streaming doesn't work:**
- Check browser console for EventSource errors
- Test SSE endpoint directly: `curl -N http://localhost:8080/api/debates/:id/stream`
- Might need to add CORS headers

**If UI looks bad:**
- Rebuild with Tailwind
- Check if CSS is loading
- Inspect elements to see if classes are applied

### Step 4: Improve UX (Based on Findings)

Possible improvements:
- [ ] Add loading spinners for API calls
- [ ] Add error boundaries for React errors
- [ ] Improve empty states
- [ ] Add toast notifications for actions
- [ ] Improve mobile layout
- [ ] Add skeleton screens while loading
- [ ] Better loading states during debate creation
- [ ] Add confirmation dialogs for delete actions

## Quick Reference Commands

```bash
# Project root
cd /Users/azuan/Workspace/Projects/dbate

# Build frontend only
cd web/app && npm run build

# Build backend only (with embedded frontend)
cd /Users/azuan/Workspace/Projects/dbate && make build-cli

# Build everything
make build

# Run server
./bin/dbate serve --port 8080

# Run frontend dev server (with hot reload)
cd web/app && npm run dev
# Then in another terminal:
make serve
```

## File Structure Reference

```
/Users/azuan/Workspace/Projects/dbate/
├── web/
│   ├── app/                        # React frontend
│   │   ├── src/
│   │   │   ├── components/        # Navigation, TurnCard
│   │   │   ├── pages/             # NewDebate, History, DebateView
│   │   │   ├── hooks/             # useDebateStream
│   │   │   ├── lib/               # api.ts
│   │   │   ├── types/             # TypeScript types
│   │   │   └── App.tsx
│   │   ├── dist/                  # Build output (embedded in Go)
│   │   ├── package.json
│   │   └── vite.config.ts
│   └── handlers/                   # Go HTTP handlers
│       ├── handlers.go            # Main routes
│       ├── streaming.go           # SSE endpoint
│       └── spa.go                 # Serves React app
├── bin/
│   └── dbate                      # Built binary
└── Makefile
```

## API Endpoints Reference

```
# REST API
GET  /api/providers          - List AI providers
GET  /api/debates            - List all debates
GET  /api/debates/:id        - Get debate + turns
POST /debates                - Create new debate
DELETE /debates/:id          - Delete debate

# Streaming
GET  /api/debates/:id/stream - SSE for real-time updates

# Export
GET  /debates/:id/export/markdown
GET  /debates/:id/export/pdf
GET  /debates/:id/export/json
```

## Testing Checklist

When you continue:
- [ ] Open http://localhost:8080 in browser
- [ ] Check browser console for errors
- [ ] Test navigation (click links, back button)
- [ ] Create a new debate
- [ ] View an existing debate
- [ ] Check if streaming works
- [ ] Test on mobile viewport
- [ ] Try dark mode (if applicable)
- [ ] Test delete debate
- [ ] Test export (if implemented)

## Known Working API Response

The API returns data in this structure:
```json
{
  "debate": {
    "id": "...",
    "topic": "...",
    "agent_a": { "id": "...", "name": "Agent A (Pragmatist)", ... },
    "agent_b": { "id": "...", "name": "Agent B (Skeptic)", ... },
    "status": "completed",
    "conclusion": { ... }
  },
  "turns": [...]
}
```

Frontend expects snake_case which matches backend, so no transformation needed.

## Debug Tips

1. **If page is blank:**
   - Check browser console
   - Check if JS file loaded (Network tab)
   - View page source - should see React root div

2. **If styles are missing:**
   - Check if CSS file loaded
   - Rebuild frontend with Tailwind

3. **If API calls fail:**
   - Check Network tab in browser
   - Verify backend is running
   - Check CORS headers if needed

4. **If routing doesn't work:**
   - Make sure Go's SPA handler is catching all routes
   - Check that RegisterSPARoutes is called last

Good luck! The infrastructure is solid, just need to debug and polish the UX.

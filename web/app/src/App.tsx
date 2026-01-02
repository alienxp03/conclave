import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Navigation } from './components/Navigation';
import { NewCouncil } from './pages/NewCouncil';
import { CouncilView } from './pages/CouncilView';
import { DebateView } from './pages/DebateView';
import { History } from './pages/History';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen flex bg-brand-bg text-[#d3c6aa]">
          <Navigation />

          <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
            <main className="flex-1 overflow-y-auto p-4 md:p-8">
              <div className="max-w-5xl mx-auto">
                <Routes>
                  <Route path="/" element={<NewCouncil />} />
                  <Route path="/councils/:id" element={<CouncilView />} />
                  <Route path="/debates/:id" element={<DebateView />} />
                  <Route path="/history" element={<History />} />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </div>
            </main>

            <footer className="border-t border-brand-border bg-brand-bg py-4 px-8">
              <div className="flex justify-between items-center">
                <p className="text-[#859289] text-xs">conclave - AI-powered council & debate tool</p>
                <div className="flex items-center gap-4 text-xs text-[#859289]">
                  <a href="/" className="hover:text-brand-primary transition-colors">New Council</a>
                  <span>•</span>
                  <a href="/history" className="hover:text-brand-primary transition-colors">History</a>
                </div>
              </div>
            </footer>
          </div>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;

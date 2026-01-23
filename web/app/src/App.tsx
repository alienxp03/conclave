import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState, useEffect } from 'react';
import { Navigation } from './components/Navigation';
import { NewCouncil } from './pages/NewCouncil';
import { CouncilView } from './pages/CouncilView';
import { DebateView } from './pages/DebateView';
import { History } from './pages/History';
import { Projects } from './pages/Projects';
import { ProjectDetail } from './pages/ProjectDetail';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  const [isCollapsed, setIsCollapsed] = useState(() => {
    const saved = localStorage.getItem('sidebar-collapsed');
    return saved === 'true';
  });

  useEffect(() => {
    localStorage.setItem('sidebar-collapsed', String(isCollapsed));
  }, [isCollapsed]);

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen flex bg-brand-bg text-[#d3c6aa] relative">
          {!isCollapsed && (
            <div
              className="md:hidden fixed inset-0 bg-black/50 z-30 transition-opacity animate-fadeIn"
              onClick={() => setIsCollapsed(true)}
            />
          )}
          <Navigation isCollapsed={isCollapsed} setIsCollapsed={setIsCollapsed} />

          {isCollapsed && (
            <button
              onClick={() => setIsCollapsed(false)}
              className="md:hidden fixed top-2 left-2 z-50 bg-brand-card p-2 rounded-md border border-brand-border text-gray-400 hover:text-white shadow-lg"
              title="Expand Sidebar"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          )}

          <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
            <main className="flex-1 overflow-y-auto p-2 md:p-8">
              <div className="max-w-5xl mx-auto">
                <Routes>
                  <Route path="/" element={<NewCouncil />} />
                  <Route path="/councils/:id" element={<CouncilView />} />
                  <Route path="/debates/:id" element={<DebateView />} />
                  <Route path="/history" element={<History />} />
                  <Route path="/projects" element={<Projects />} />
                  <Route path="/projects/:id" element={<ProjectDetail />} />
                  <Route path="*" element={<Navigate to="/" replace />} />
                </Routes>
              </div>
            </main>

            <footer className="border-t border-brand-border bg-brand-bg py-4 px-4 md:px-8">
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

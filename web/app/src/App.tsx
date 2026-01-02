import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Navigation } from './components/Navigation';
import { NewCouncil } from './pages/NewCouncil';
import { CouncilView } from './pages/CouncilView';
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
        <div className="min-h-screen flex flex-col bg-gray-900">
          <Navigation />

          <main className="flex-1 max-w-7xl w-full mx-auto py-6 px-4 sm:px-6 lg:px-8">
            <Routes>
              <Route path="/" element={<NewCouncil />} />
              <Route path="/councils/:id" element={<CouncilView />} />
              <Route path="/history" element={<History />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </main>

          <footer className="border-t border-gray-700 bg-gray-800 mt-auto">
            <div className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
              <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
                <p className="text-gray-400 text-sm">dbate - AI-powered debate tool</p>
                <div className="flex items-center gap-4 text-xs text-gray-500">
                  <a href="/" className="hover:text-gray-300 transition-colors">
                    New Debate
                  </a>
                  <span>•</span>
                  <a href="/history" className="hover:text-gray-300 transition-colors">
                    History
                  </a>
                </div>
              </div>
            </div>
          </footer>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;

import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';

export function History() {
  const { data: debates, isLoading: loadingDebates } = useQuery({
    queryKey: ['debates'],
    queryFn: () => api.getDebates(50, 0),
  });

  const { data: councils, isLoading: loadingCouncils } = useQuery({
    queryKey: ['councils'],
    queryFn: () => api.getCouncils(50, 0),
  });

  const isLoading = loadingDebates || loadingCouncils;

  // Merge and sort by date
  const allItems = [
    ...(debates || []).map(d => ({ ...d, type: 'debate' as const })),
    ...(councils || []).map(c => ({ 
      id: c.id, 
      title: c.title,
      topic: c.topic, 
      cwd: c.cwd,
      status: c.status, 
      created_at: c.created_at, 
      type: 'council' as const,
      member_count: c.member_count
    }))
  ].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

  const statusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-brand-primary/10 text-brand-primary';
      case 'in_progress':
        return 'bg-brand-blue/10 text-brand-blue';
      case 'failed':
        return 'bg-brand-accent/10 text-brand-accent';
      default:
        return 'bg-brand-border text-[#859289]';
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="inline-block p-6 bg-brand-card rounded-2xl border-2 border-brand-primary shadow-2xl">
            <svg
              className="animate-spin h-16 w-16 text-brand-primary mx-auto"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            <p className="mt-4 text-gray-400">Loading history...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">Your Councils</h1>
          <p className="text-gray-400">Browse and revisit your AI councils and debates</p>
        </div>
        <Link
          to="/"
          className="bg-brand-primary hover:bg-[#b8cc95] text-[#2b3339] px-6 py-3 rounded-xl text-base font-bold inline-flex items-center shadow-lg transform transition-all duration-200 hover:scale-105"
        >
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 4v16m8-8H4"
            />
          </svg>
          New Council
        </Link>
      </div>

      {!allItems || allItems.length === 0 ? (
        <div className="text-center py-20 bg-brand-card rounded-2xl border border-brand-border">
          <div className="inline-block p-6 bg-brand-bg rounded-full mb-6">
            <svg
              className="h-20 w-20 text-gray-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
          </div>
          <h3 className="text-2xl font-semibold text-white mb-2">No history yet</h3>
          <p className="text-gray-400 mb-8 max-w-md mx-auto">
            Start your first AI-powered discussion and watch different perspectives unfold
          </p>
          <Link
            to="/"
            className="inline-flex items-center px-8 py-4 bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 text-white rounded-xl font-semibold shadow-lg transform transition-all duration-200 hover:scale-105"
          >
            <svg className="w-6 h-6 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Start New Council
          </Link>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {allItems.map((item) => (
            <Link 
              key={item.id} 
              to={item.type === 'debate' ? `/debates/${item.id}` : `/councils/${item.id}`} 
              className="block group"
            >
              <div className="bg-brand-card border border-brand-border rounded-xl p-6 hover:border-brand-primary hover:shadow-xl transition-all duration-200 h-full flex flex-col">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex flex-col gap-1">
                    <span className={`px-3 py-1 rounded-full text-xs font-semibold ${statusColor(item.status)}`}>
                      {item.status}
                    </span>
                    <span className="text-[10px] uppercase tracking-wider text-gray-500 font-bold px-1">
                      {item.type}
                    </span>
                  </div>
                  {item.type === 'debate' && (item as any).read_only && (
                    <span className="text-yellow-400" title="Read-only">
                      🔒
                    </span>
                  )}
                </div>

                <h3 className="text-lg font-semibold text-white mb-4 group-hover:text-brand-primary transition-colors line-clamp-2">
                  {item.title || item.topic}
                </h3>

                <div className="mt-auto space-y-3">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-400">
                      {item.type === 'debate' ? 'Turns' : 'Members'}
                    </span>
                    <span className="text-white font-medium">
                      {item.type === 'debate' ? (item as any).turn_count : (item as any).member_count}
                    </span>
                  </div>

                  <div className="flex items-center gap-2 text-xs text-gray-500">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    {new Date(item.created_at).toLocaleString()}
                  </div>

                  {item.type === 'debate' && (
                    <div className="pt-3 border-t border-brand-border">
                      <div className="flex items-center justify-between text-xs">
                        <span className="text-brand-primary font-bold">{(item as any).agent_a}</span>
                        <span className="text-[#5c6a72]">vs</span>
                        <span className="text-brand-secondary font-bold">{(item as any).agent_b}</span>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

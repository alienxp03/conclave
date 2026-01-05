import { Link, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';

export interface NavigationProps {
  isCollapsed: boolean;
  setIsCollapsed: (collapsed: boolean) => void;
}

export function Navigation({ isCollapsed, setIsCollapsed }: NavigationProps) {
  const location = useLocation();

  const { data: debates } = useQuery({
    queryKey: ['debates'],
    queryFn: () => api.getDebates(20, 0),
    refetchInterval: 3000,
  });

  const { data: councils } = useQuery({
    queryKey: ['councils'],
    queryFn: () => api.getCouncils(20, 0),
    refetchInterval: 3000,
  });

  const allItems = [
    ...(debates || []).map(d => ({ ...d, type: 'debate' as const })),
    ...(councils || []).map(c => ({
      id: c.id,
      title: c.title,
      topic: c.topic,
      created_at: c.created_at,
      type: 'council' as const,
    }))
  ].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

  return (
    <div className={`
      ${isCollapsed ? 'w-0 md:w-16 -translate-x-full md:translate-x-0' : 'w-64 translate-x-0'} 
      flex-shrink-0 bg-brand-card border-r border-brand-border flex flex-col h-screen overflow-hidden transition-all duration-300 fixed md:relative z-40
    `}>
      <div className={`p-4 border-b border-brand-border flex items-center justify-between ${isCollapsed ? 'opacity-0 md:opacity-100' : 'opacity-100'}`}>
        {!isCollapsed ? (
          <>
            <Link to="/" className="flex items-center space-x-2 group">
              <span className="text-brand-primary transform transition-transform group-hover:scale-110">
                <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 13.5997 2.37562 15.1116 3.04348 16.4522L2 22L7.54777 20.9565C8.88837 21.6244 10.4003 22 12 22Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M8 12H8.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M12 12H12.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M16 12H16.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </span>
              <span className="font-bold text-base text-white">conclave</span>
            </Link>
            <div className="flex items-center gap-1">
              <Link
                to="/"
                className="text-gray-400 hover:text-white p-1 rounded-md hover:bg-brand-bg transition-colors"
                title="New Conversation"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              </Link>
              <button
                onClick={() => setIsCollapsed(true)}
                className="text-gray-400 hover:text-white p-1 rounded-md hover:bg-brand-bg transition-colors"
                title="Collapse Sidebar"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
                </svg>
              </button>
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center gap-3 w-full">
            <Link to="/" className="group" title="Conclave">
              <span className="text-brand-primary transform transition-transform group-hover:scale-110">
                <svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 13.5997 2.37562 15.1116 3.04348 16.4522L2 22L7.54777 20.9565C8.88837 21.6244 10.4003 22 12 22Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M8 12H8.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M12 12H12.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M16 12H16.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </span>
            </Link>
            <Link
              to="/"
              className="text-gray-400 hover:text-white p-1 rounded-md hover:bg-brand-bg transition-colors"
              title="New Conversation"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </Link>
            <button
              onClick={() => setIsCollapsed(false)}
              className="text-gray-400 hover:text-white p-1 rounded-md hover:bg-brand-bg transition-colors"
              title="Expand Sidebar"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {!isCollapsed && (
          <div className="px-2 py-2 text-[10px] font-bold text-[#859289] uppercase tracking-widest opacity-70">
            Recent Conversations
          </div>
        )}
        {allItems.map((item) => {
          const isActive = location.pathname.includes(item.id);
          return (
            <Link
              key={item.id}
              to={item.type === 'debate' ? `/debates/${item.id}` : `/councils/${item.id}`}
              className={`block rounded-lg text-sm transition-all duration-200 ${
                isCollapsed ? 'px-2 py-2' : 'px-3 py-2'
              } ${
                isActive
                  ? 'bg-brand-primary text-[#2b3339] shadow-lg font-bold'
                  : 'text-[#9da9a0] hover:text-[#d3c6aa] hover:bg-brand-bg'
              }`}
              title={item.title || item.topic}
            >
              {isCollapsed ? (
                <div className="flex justify-center">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                  </svg>
                </div>
              ) : (
                item.title || "New conversation"
              )}
            </Link>
          );
        })}
        {allItems.length === 0 && !isCollapsed && (
          <div className="px-3 py-8 text-center text-xs text-[#859289]">
            No history yet
          </div>
        )}
      </div>

      {!isCollapsed && (
        <div className="p-4 border-t border-brand-border">
          <Link
            to="/history"
            className="flex items-center space-x-2 text-sm text-[#859289] hover:text-brand-primary transition-colors"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>View all history</span>
          </Link>
        </div>
      )}
    </div>
  );
}

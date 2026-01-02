import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { useDebateStream } from '../hooks/useDebateStream';
import { TurnCard } from '../components/TurnCard';

export function DebateView() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [deleting, setDeleting] = useState(false);

  const { data, isLoading, error } = useQuery({
    queryKey: ['debate', id],
    queryFn: () => api.getDebate(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const debate = query.state.data?.debate;
      return debate?.status === 'in_progress' ? 5000 : false;
    },
  });

  const { turns: streamingTurns, streamingTurn, debate: streamedDebate } = useDebateStream(
    data?.debate.status === 'in_progress' ? id : undefined,
    data?.turns
  );

  const debate = streamedDebate || data?.debate;
  const turns = streamingTurns.length > 0 ? streamingTurns : data?.turns || [];

  const handleDelete = async () => {
    if (!id || !confirm('Are you sure you want to delete this debate?')) return;

    setDeleting(true);
    try {
      await api.deleteDebate(id);
      navigate('/history');
    } catch (err) {
      alert('Failed to delete debate');
      setDeleting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="inline-block p-6 bg-gray-800 rounded-2xl border-2 border-blue-500 shadow-2xl">
            <svg
              className="animate-spin h-16 w-16 text-blue-500 mx-auto"
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
            <p className="mt-4 text-gray-400">Loading debate...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error || !debate) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-white mb-4">Debate not found</h2>
          <Link to="/" className="text-blue-400 hover:text-blue-300">
            Go home
          </Link>
        </div>
      </div>
    );
  }

  const statusColor = {
    completed: 'bg-green-100 text-green-800',
    in_progress: 'bg-blue-100 text-blue-800',
    failed: 'bg-red-100 text-red-800',
    pending: 'bg-gray-100 text-gray-800',
  }[debate.status];

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="mb-4" aria-label="Breadcrumb">
        <ol className="flex items-center space-x-2 text-sm text-gray-400">
          <li>
            <Link to="/" className="hover:text-white transition-colors">
              Home
            </Link>
          </li>
          <li>
            <span className="text-gray-600">/</span>
          </li>
          <li>
            <Link to="/history" className="hover:text-white transition-colors">
              History
            </Link>
          </li>
          <li>
            <span className="text-gray-600">/</span>
          </li>
          <li className="text-gray-300">
            {debate.topic.substring(0, 30)}
            {debate.topic.length > 30 ? '...' : ''}
          </li>
        </ol>
      </nav>

      {/* Header */}
      <div className="bg-gray-800 shadow-xl rounded-xl p-8 border border-gray-700">
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <div className="flex items-center space-x-3">
              <h1 className="text-2xl font-bold text-white">{debate.topic}</h1>
              {debate.read_only && (
                <span className="text-yellow-400" title="Read-only">
                  🔒
                </span>
              )}
            </div>
            <div className="mt-3 flex items-center space-x-4 text-sm text-gray-400">
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColor}`}>
                {debate.status}
              </span>
              <span>
                {turns.length}/{debate.total_turns} turns
              </span>
            </div>
          </div>
          <div className="flex space-x-2">
            {debate.status === 'completed' && (
              <div className="relative group">
                <button className="inline-flex items-center px-3 py-2 border border-gray-600 text-sm font-medium rounded-md text-gray-300 bg-gray-700 hover:bg-gray-600">
                  Export
                  <svg
                    className="ml-2 h-4 w-4"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>
                <div className="hidden group-hover:block absolute right-0 mt-2 w-48 bg-gray-700 rounded-md shadow-lg z-10 border border-gray-600">
                  <a
                    href={`/debates/${debate.id}/export/markdown`}
                    className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-600"
                  >
                    📝 Markdown
                  </a>
                  <a
                    href={`/debates/${debate.id}/export/pdf`}
                    className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-600"
                  >
                    📄 PDF
                  </a>
                  <a
                    href={`/debates/${debate.id}/export/json`}
                    className="block px-4 py-2 text-sm text-gray-300 hover:bg-gray-600"
                  >
                    📊 JSON
                  </a>
                </div>
              </div>
            )}

            {!debate.read_only && (
              <button
                onClick={handleDelete}
                disabled={deleting}
                className="inline-flex items-center px-3 py-2 border border-red-600 text-sm font-medium rounded-md text-red-400 bg-gray-700 hover:bg-red-900 hover:text-red-300 disabled:opacity-50"
              >
                {deleting ? 'Deleting...' : 'Delete'}
              </button>
            )}
          </div>
        </div>

        {/* Agents */}
        <div className="mt-6 grid grid-cols-2 gap-4">
          <div className="border border-blue-700 rounded-lg p-4 bg-blue-900 bg-opacity-30">
            <div className="font-medium text-blue-300 text-lg">{debate.agent_a.name}</div>
            <div className="text-sm text-gray-400 mt-1">{debate.agent_a.persona}</div>
          </div>
          <div className="border border-green-700 rounded-lg p-4 bg-green-900 bg-opacity-30">
            <div className="font-medium text-green-300 text-lg">{debate.agent_b.name}</div>
            <div className="text-sm text-gray-400 mt-1">{debate.agent_b.persona}</div>
          </div>
        </div>
      </div>

      {/* Turns */}
      {turns.length === 0 && !streamingTurn ? (
        <div className="animate-fadeIn bg-gradient-to-br from-gray-800 to-gray-900 shadow-2xl rounded-2xl p-20 text-center border border-gray-700 relative overflow-hidden">
          <div className="absolute inset-0 opacity-10">
            <div className="absolute top-10 left-10 w-32 h-32 bg-blue-500 rounded-full blur-3xl" />
            <div className="absolute bottom-10 right-10 w-32 h-32 bg-green-500 rounded-full blur-3xl" />
          </div>
          <div className="relative z-10">
            <div className="inline-block p-6 bg-gray-900 rounded-full mb-6 shadow-lg">
              <svg
                className="h-20 w-20 text-blue-400 animate-pulse"
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
            <h3 className="text-3xl font-bold text-white mb-3">Agents are debating...</h3>
            <p className="text-gray-400 text-lg max-w-md mx-auto">
              AI agents are analyzing the topic and formulating their arguments.
            </p>
            <div className="mt-8 flex justify-center gap-2">
              <div className="w-3 h-3 bg-blue-500 rounded-full animate-bounce" />
              <div
                className="w-3 h-3 bg-blue-400 rounded-full animate-bounce"
                style={{ animationDelay: '0.15s' }}
              />
              <div
                className="w-3 h-3 bg-blue-300 rounded-full animate-bounce"
                style={{ animationDelay: '0.3s' }}
              />
            </div>
          </div>
        </div>
      ) : (
        <div className="bg-gradient-to-br from-gray-800 to-gray-900 rounded-2xl border border-gray-700 overflow-hidden">
          <div className="bg-gray-900 bg-opacity-50 px-8 py-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold text-white flex items-center gap-2">
              <svg
                className="w-6 h-6 text-blue-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
                />
              </svg>
              Debate Transcript
            </h2>
          </div>

          <div className="p-6 space-y-6 max-h-[800px] overflow-y-auto">
            {turns.map((turn) => (
              <TurnCard key={turn.id} turn={turn} debate={debate} />
            ))}
            {streamingTurn && (
              <TurnCard
                turn={{
                  id: 'streaming',
                  debate_id: debate.id,
                  agent_id: streamingTurn.agentId,
                  number: streamingTurn.number,
                  content: streamingTurn.content,
                  created_at: new Date().toISOString(),
                }}
                debate={debate}
                isStreaming
              />
            )}
          </div>
        </div>
      )}

      {/* Conclusion */}
      {debate.conclusion && (
        <div
          className={`bg-gradient-to-br from-gray-800 via-gray-850 to-gray-900 shadow-2xl rounded-2xl p-10 border-2 ${
            debate.conclusion.agreed ? 'border-green-500' : 'border-yellow-500'
          } relative overflow-hidden`}
        >
          <div
            className={`absolute top-0 right-0 w-64 h-64 ${
              debate.conclusion.agreed ? 'bg-green-500' : 'bg-yellow-500'
            } opacity-5 rounded-full blur-3xl`}
          />

          <div className="relative z-10">
            <div className="flex items-center gap-4 mb-8">
              <div
                className={`w-16 h-16 rounded-full ${
                  debate.conclusion.agreed ? 'bg-green-500' : 'bg-yellow-500'
                } bg-opacity-20 flex items-center justify-center`}
              >
                <span className="text-4xl">{debate.conclusion.agreed ? '🤝' : '⚔️'}</span>
              </div>
              <div>
                <h2 className="text-3xl font-bold text-white">
                  {debate.conclusion.agreed
                    ? debate.conclusion.early_consensus
                      ? 'Consensus Reached Early!'
                      : 'Consensus Reached'
                    : 'No Consensus'}
                </h2>
                <p className="text-gray-400 mt-1">
                  {debate.conclusion.agreed
                    ? 'Both agents found common ground'
                    : 'Agents maintained different viewpoints'}
                </p>
              </div>
            </div>

            <div className="bg-gray-900 bg-opacity-30 rounded-xl p-8 border border-gray-700">
              <h3 className="text-xl font-semibold text-white mb-4 flex items-center gap-2">
                <svg
                  className="w-6 h-6 text-blue-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                Summary
              </h3>
              <p className="text-gray-200 text-lg leading-relaxed">{debate.conclusion.summary}</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

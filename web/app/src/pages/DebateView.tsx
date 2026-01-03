import { useState, useRef, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { api } from '../lib/api';
import { useDebateStream } from '../hooks/useDebateStream';
import { TurnCard } from '../components/TurnCard';
import type { Turn, Conclusion } from '../types';

export function DebateView() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [deleting, setDeleting] = useState(false);
  const [followUp, setFollowUp] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['debate', id],
    queryFn: () => api.getDebate(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const debate = query.state.data?.debate;
      return debate?.status === 'in_progress' ? 3000 : false;
    },
  });

  const { turns: streamingTurns, streamingTurn, debate: streamedDebate } = useDebateStream(
    data?.debate.status === 'in_progress' ? id : undefined,
    data?.turns
  );

  const debate = streamedDebate || data?.debate;
  const turns = streamingTurns.length > 0 ? streamingTurns : data?.turns || [];

  const followUpMutation = useMutation({
    mutationFn: (content: string) => api.addDebateFollowUp(id!, content),
    onSuccess: () => {
      setFollowUp('');
      queryClient.invalidateQueries({ queryKey: ['debate', id] });
    },
  });

  // Auto-scroll to bottom on new turns
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [turns.length, streamingTurn?.content]);

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

  const handleFollowUp = (e: React.FormEvent) => {
    e.preventDefault();
    if (!followUp.trim() || followUpMutation.isPending) return;
    followUpMutation.mutate(followUp.trim());
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
    completed: 'bg-[#a7c080]/10 text-[#a7c080]',
    in_progress: 'bg-[#7fbbb3]/10 text-[#7fbbb3]',
    failed: 'bg-[#e67e80]/10 text-[#e67e80]',
    pending: 'bg-[#3a454a] text-[#859289]',
  }[debate.status];

  // Group turns and conclusions by round
  const rounds = Array.from(new Set([...turns.map(t => t.round), ...(debate.conclusions?.map(c => c.round) || [])])).sort((a, b) => a - b);
  
  const turnsByRound = turns.reduce((acc, turn) => {
    if (!acc[turn.round]) acc[turn.round] = [];
    acc[turn.round].push(turn);
    return acc;
  }, {} as Record<number, Turn[]>);

  const conclusionsByRound = (debate.conclusions || []).reduce((acc, conclusion) => {
    acc[conclusion.round] = conclusion;
    return acc;
  }, {} as Record<number, Conclusion>);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-brand-card shadow-xl rounded-xl p-8 border border-brand-border">
        <div className="flex justify-between items-start">
          <div className="flex-1">
            <div className="flex items-center space-x-3">
              <h1 className="text-xl font-bold text-[#d3c6aa]">{debate.title || debate.topic}</h1>
              {debate.read_only && (
                <span className="text-brand-secondary" title="Read-only">
                  🔒
                </span>
              )}
            </div>
            {debate.title && (
              <p className="text-[#859289] mt-1">{debate.topic}</p>
            )}
            <div className="mt-3 flex items-center space-x-4 text-sm text-[#859289]">
              <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColor}`}>
                {debate.status}
              </span>
              <span>
                {rounds.length} Round{rounds.length !== 1 ? 's' : ''} • {turns.length} Turn{turns.length !== 1 ? 's' : ''}
              </span>
            </div>
          </div>
          <div className="flex space-x-2">
            {debate.status === 'completed' && (
              <div className="relative group">
                <button className="inline-flex items-center px-3 py-2 border border-brand-border text-sm font-medium rounded-md text-[#d3c6aa] bg-brand-card hover:bg-brand-border">
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
                <div className="hidden group-hover:block absolute right-0 mt-2 w-48 bg-brand-card rounded-md shadow-lg z-10 border border-brand-border">
                  <a
                    href={`/debates/${debate.id}/export/markdown`}
                    className="block px-4 py-2 text-sm text-[#d3c6aa] hover:bg-brand-bg"
                  >
                    📝 Markdown
                  </a>
                  <a
                    href={`/debates/${debate.id}/export/pdf`}
                    className="block px-4 py-2 text-sm text-[#d3c6aa] hover:bg-brand-bg"
                  >
                    📄 PDF
                  </a>
                  <a
                    href={`/debates/${debate.id}/export/json`}
                    className="block px-4 py-2 text-sm text-[#d3c6aa] hover:bg-brand-bg"
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
                className="inline-flex items-center px-3 py-2 border border-brand-accent text-sm font-medium rounded-md text-brand-accent bg-brand-card hover:bg-brand-accent hover:text-brand-bg disabled:opacity-50"
              >
                {deleting ? 'Deleting...' : 'Delete'}
              </button>
            )}
          </div>
        </div>

        {/* Agents */}
        <div className="mt-6 grid grid-cols-2 gap-4">
          <div className="border border-brand-primary border-opacity-30 rounded-lg p-4 bg-brand-primary bg-opacity-5">
            <div className="font-medium text-brand-primary text-lg">{debate.agent_a.name}</div>
            <div className="text-sm text-[#859289] mt-1">{debate.agent_a.persona}</div>
          </div>
          <div className="border border-brand-secondary border-opacity-30 rounded-lg p-4 bg-brand-secondary bg-opacity-5">
            <div className="font-medium text-brand-secondary text-lg">{debate.agent_b.name}</div>
            <div className="text-sm text-[#859289] mt-1">{debate.agent_b.persona}</div>
          </div>
        </div>

        {debate.cwd && (
          <div className="mt-6 flex items-center text-[#d3c6aa] text-xs font-mono bg-brand-bg bg-opacity-40 px-3 py-1.5 rounded border border-brand-border w-fit shadow-inner">
            <svg className="w-3.5 h-3.5 mr-1.5 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            <span className="opacity-60 mr-1 text-[#859289]">Session dir:</span> {debate.cwd}
          </div>
        )}
      </div>

      {/* Resume Status Indicator */}
      {debate.status === 'in_progress' && rounds.length > 1 && (
        <div className="animate-fadeIn bg-gradient-to-r from-blue-500/10 via-brand-primary/10 to-blue-500/10 border-2 border-blue-500/30 rounded-xl p-5 shadow-lg">
          <div className="flex items-center gap-4">
            <div className="relative">
              <div className="w-3 h-3 bg-blue-500 rounded-full animate-ping absolute"></div>
              <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
            </div>
            <div className="flex-1">
              <div className="text-blue-400 font-bold text-lg">Resuming Deliberation</div>
              <div className="text-sm text-[#859289]">
                Round {Math.max(...rounds)} in progress - agents are responding to your follow-up...
              </div>
            </div>
            <div className="flex items-center gap-2 text-xs text-[#859289] bg-brand-bg/50 px-3 py-1.5 rounded-full">
              <svg className="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              Processing
            </div>
          </div>
        </div>
      )}

      {/* Consensus Evolution Timeline */}
      {debate.conclusions && debate.conclusions.length > 1 && debate.status === 'completed' && (
        <div className="animate-fadeIn bg-gradient-to-br from-brand-card via-brand-card to-brand-bg rounded-2xl border border-brand-border p-8 shadow-xl">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 bg-brand-primary/20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <div>
              <h3 className="text-lg font-bold text-[#d3c6aa]">Consensus Evolution</h3>
              <p className="text-sm text-[#859289]">Track how the agents' positions evolved across {debate.conclusions.length} rounds</p>
            </div>
          </div>

          <div className="space-y-3">
            {debate.conclusions.map((conclusion, idx) => (
              <div key={idx} className="flex items-start gap-4 group">
                <div className="flex flex-col items-center">
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm border-2 ${
                    conclusion.agreed
                      ? 'bg-brand-primary/20 text-brand-primary border-brand-primary'
                      : 'bg-brand-secondary/20 text-brand-secondary border-brand-secondary'
                  }`}>
                    R{conclusion.round}
                  </div>
                  {idx < (debate.conclusions?.length || 0) - 1 && (
                    <div className="w-0.5 h-full min-h-[40px] bg-brand-border mt-2"></div>
                  )}
                </div>

                <div className={`flex-1 p-5 rounded-xl border-2 transition-all group-hover:shadow-lg ${
                  conclusion.agreed
                    ? 'border-brand-primary/30 bg-brand-primary/5 group-hover:bg-brand-primary/10'
                    : 'border-brand-secondary/30 bg-brand-secondary/5 group-hover:bg-brand-secondary/10'
                }`}>
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <span className="text-lg">{conclusion.agreed ? '🤝' : '⚔️'}</span>
                      <span className="text-xs font-bold uppercase tracking-wider">
                        {conclusion.agreed ? 'Consensus Reached' : 'Divergent Views'}
                      </span>
                      {conclusion.early_consensus && (
                        <span className="text-xs bg-brand-primary/20 text-brand-primary px-2 py-0.5 rounded-full">
                          Early Consensus
                        </span>
                      )}
                    </div>
                    {idx === (debate.conclusions?.length || 0) - 1 && (
                      <span className="text-xs bg-brand-bg px-2 py-1 rounded-full text-brand-primary border border-brand-primary/30">
                        Latest
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-[#d3c6aa] line-clamp-3 group-hover:line-clamp-none transition-all">
                    {conclusion.summary}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Rounds and Transcript */}
      <div className="space-y-12">
        {rounds.map((roundNum) => {
          const roundConclusion = conclusionsByRound[roundNum];
          const roundTurns = turnsByRound[roundNum] || [];
          const userTurn = roundTurns.find(t => t.agent_id === 'user');

          return (
            <div key={roundNum} className="space-y-6">
              <div className="flex items-center gap-4">
                <div className="h-px flex-1 bg-brand-border" />
                <div className="flex items-center gap-3">
                  <span className="text-xs font-bold text-[#859289] uppercase tracking-widest">
                    Round {roundNum}
                  </span>
                  {roundConclusion && (
                    <span className={`text-xs px-2 py-1 rounded-full font-medium ${
                      roundConclusion.agreed
                        ? 'bg-brand-primary/20 text-brand-primary'
                        : 'bg-brand-secondary/20 text-brand-secondary'
                    }`}>
                      {roundConclusion.agreed ? '✓ Consensus' : '• Divergent'}
                    </span>
                  )}
                  <span className="text-xs text-[#859289] bg-brand-bg px-2 py-0.5 rounded-full">
                    {roundTurns.filter(t => t.agent_id !== 'user').length} turns
                  </span>
                </div>
                <div className="h-px flex-1 bg-brand-border" />
              </div>

              {/* User Follow-up Directive */}
              {userTurn && (
                <div className="animate-fadeIn bg-gradient-to-r from-brand-primary/10 to-transparent border-l-4 border-brand-primary p-6 rounded-r-xl shadow-inner">
                  <div className="flex items-center gap-3 mb-3">
                    <div className="w-8 h-8 bg-brand-primary/20 rounded-lg flex items-center justify-center">
                      <svg className="w-5 h-5 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                      </svg>
                    </div>
                    <span className="text-xs font-bold text-brand-primary uppercase tracking-wider">Your Follow-up</span>
                  </div>
                  <p className="text-[#d3c6aa] text-base italic pl-11">"{userTurn.content}"</p>
                </div>
              )}

              <div className="space-y-6">
                {roundTurns.map((turn) => (
                  <TurnCard key={turn.id} turn={turn} debate={debate} />
                ))}

                {/* Show streaming turn if it belongs to this round */}
                {streamingTurn && streamingTurn.round === roundNum && (
                  <TurnCard
                    turn={{
                      id: 'streaming',
                      debate_id: debate.id,
                      agent_id: streamingTurn.agentId,
                      number: streamingTurn.number,
                      round: streamingTurn.round,
                      content: streamingTurn.content,
                      created_at: new Date().toISOString(),
                    }}
                    debate={debate}
                    isStreaming
                  />
                )}

                {/* Round Conclusion */}
                {conclusionsByRound[roundNum] && (
                  <div
                    className={`animate-fadeIn bg-gradient-to-br from-brand-card via-brand-card to-brand-bg shadow-2xl rounded-2xl p-8 border-2 ${
                      conclusionsByRound[roundNum].agreed ? 'border-brand-primary' : 'border-brand-secondary'
                    } relative overflow-hidden`}
                  >
                    <div className="relative z-10">
                      <div className="flex items-center gap-4 mb-6">
                        <div
                          className={`w-12 h-12 rounded-full ${
                            conclusionsByRound[roundNum].agreed ? 'bg-brand-primary' : 'bg-brand-secondary'
                          } bg-opacity-20 flex items-center justify-center`}
                        >
                          <span className="text-2xl">{conclusionsByRound[roundNum].agreed ? '🤝' : '⚔️'}</span>
                        </div>
                        <div>
                          <h2 className="text-xl font-bold text-[#d3c6aa]">
                            Consensus Round {roundNum}
                          </h2>
                          <p className="text-sm text-[#859289]">
                            {conclusionsByRound[roundNum].agreed
                              ? 'Reached agreement'
                              : 'Different viewpoints maintained'}
                          </p>
                        </div>
                      </div>

                      <div className="bg-brand-bg bg-opacity-30 rounded-xl p-6 border border-brand-border">
                        <div className="text-[#d3c6aa] prose prose-invert max-w-none prose-headings:text-[#d3c6aa] prose-strong:text-brand-secondary text-sm">
                          <ReactMarkdown remarkPlugins={[remarkGfm]}>
                            {conclusionsByRound[roundNum].summary}
                          </ReactMarkdown>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          );
        })}

        {/* Empty state while starting */}
        {rounds.length === 0 && !streamingTurn && (
          <div className="animate-fadeIn bg-gradient-to-br from-brand-card to-brand-bg shadow-2xl rounded-2xl p-20 text-center border border-brand-border relative overflow-hidden">
            <div className="absolute inset-0 opacity-10">
              <div className="absolute top-10 left-10 w-32 h-32 bg-brand-primary rounded-full blur-3xl" />
              <div className="absolute bottom-10 right-10 w-32 h-32 bg-brand-secondary rounded-full blur-3xl" />
            </div>
            <div className="relative z-10">
              <div className="inline-block p-6 bg-brand-bg rounded-full mb-6 shadow-lg">
                <svg
                  className="h-20 w-20 text-brand-primary animate-pulse"
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
            </div>
          </div>
        )}

        {/* Follow-up Input */}
        {debate.status === 'completed' && !debate.read_only && (
          <div className="animate-fadeIn bg-brand-card rounded-2xl border border-brand-border p-8 shadow-2xl">
            <div className="flex items-center gap-3 mb-6">
              <div className="w-10 h-10 bg-brand-primary bg-opacity-20 rounded-lg flex items-center justify-center">
                <svg className="w-6 h-6 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-bold text-[#d3c6aa]">Push the Deliberation Further</h3>
                <p className="text-sm text-[#859289]">Add a follow-up question or directive to start another round.</p>
              </div>
            </div>

            <form onSubmit={handleFollowUp} className="space-y-4">
              <textarea
                value={followUp}
                onChange={(e) => setFollowUp(e.target.value)}
                placeholder="e.g., Consider the environmental impact as well..."
                rows={3}
                className="w-full bg-brand-bg border-2 border-brand-border rounded-xl p-4 text-[#d3c6aa] focus:border-brand-primary outline-none transition-all"
              />
              <div className="flex justify-end">
                <button
                  type="submit"
                  disabled={!followUp.trim() || followUpMutation.isPending}
                  className="px-6 py-3 bg-brand-primary hover:bg-[#b8cc95] text-[#2b3339] font-bold rounded-lg transition-all transform active:scale-95 disabled:opacity-50"
                >
                  {followUpMutation.isPending ? 'Resuming...' : 'Resume Deliberation'}
                </button>
              </div>
            </form>
          </div>
        )}
      </div>
      <div ref={scrollRef} />
    </div>
  );
}

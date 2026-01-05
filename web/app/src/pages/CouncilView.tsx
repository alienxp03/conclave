import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';
import { Message } from '../components/Message';
import { RoundContainer } from '../components/RoundContainer';
import type { CouncilResponse, CouncilRanking, CouncilSynthesis } from '../types';

// Helper functions for formatting metadata
function formatTokens(count: number): string {
  if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}k`;
  }
  return count.toString();
}

function formatDuration(ms: number): string {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  return `${ms}ms`;
}

function buildMetadata(item: { input_tokens?: number; output_tokens?: number; duration_ms?: number }): string | undefined {
  const parts: string[] = [];
  if (item.input_tokens || item.output_tokens) {
    parts.push(`↑${formatTokens(item.input_tokens || 0)} ↓${formatTokens(item.output_tokens || 0)}`);
  }
  if (item.duration_ms && item.duration_ms > 0) {
    parts.push(formatDuration(item.duration_ms));
  }
  return parts.length > 0 ? parts.join(' • ') : undefined;
}

export function CouncilView() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const [followUp, setFollowUp] = useState('');

  // Streaming state
  const [streamData, setStreamData] = useState<{
    responses: Record<string, CouncilResponse>;
    rankings: Record<string, CouncilRanking>;
    syntheses: Record<number, CouncilSynthesis>;
  }>({
    responses: {},
    rankings: {},
    syntheses: {},
  });

  const { data, isLoading, error } = useQuery({
    queryKey: ['council', id],
    queryFn: () => api.getCouncil(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.council.status;
      return status === 'in_progress' ? 3000 : false;
    },
  });

  const followUpMutation = useMutation({
    mutationFn: (content: string) => api.addCouncilFollowUp(id!, content),
    onSuccess: () => {
      setFollowUp('');
      queryClient.invalidateQueries({ queryKey: ['council', id] });
    },
  });

  // Streaming effect
  useEffect(() => {
    if (!id || data?.council.status !== 'in_progress') return;

    const eventSource = api.createCouncilStream(id);

    eventSource.addEventListener('response_collected', (e: any) => {
      const payload = JSON.parse(e.data);
      setStreamData(prev => ({
        ...prev,
        responses: { ...prev.responses, [`${payload.round}-${payload.agent_id}`]: payload }
      }));
    });

    eventSource.addEventListener('ranking_collected', (e: any) => {
      const payload = JSON.parse(e.data);
      setStreamData(prev => ({
        ...prev,
        rankings: { ...prev.rankings, [`${payload.round}-${payload.agent_id}`]: payload }
      }));
    });

    eventSource.addEventListener('synthesis_complete', (e: any) => {
      const payload = JSON.parse(e.data);
      // Payload for synthesis streaming might not have round yet, default to latest
      setStreamData(prev => {
        const validRounds = Object.values(prev.responses)
          .map(r => r.round)
          .filter(r => typeof r === 'number' && !isNaN(r) && r > 0);
        const round = payload.round || (validRounds.length > 0 ? Math.max(...validRounds) : 1);

        // Only add synthesis if we have a valid round number
        if (typeof round !== 'number' || isNaN(round) || round <= 0) {
          return prev;
        }

        return {
          ...prev,
          syntheses: { ...prev.syntheses, [round]: { round, content: payload.synthesis, created_at: new Date().toISOString() } }
        };
      });
    });

    eventSource.addEventListener('complete', () => {
      queryClient.invalidateQueries({ queryKey: ['council', id] });
      eventSource.close();
    });

    eventSource.addEventListener('error', (e) => {
      console.error('Stream error', e);
      eventSource.close();
    });

    return () => eventSource.close();
  }, [id, data?.council.status, queryClient]);

  const handleFollowUp = (e: React.FormEvent) => {
    e.preventDefault();
    if (!followUp.trim() || followUpMutation.isPending) return;
    followUpMutation.mutate(followUp.trim());
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleFollowUp(e as any);
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white"></div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="text-center py-12">
        <h2 className="text-2xl font-bold text-red-400 mb-4">Error loading council</h2>
        <p className="text-gray-400">{error instanceof Error ? error.message : 'Unknown error'}</p>
        <button
          onClick={() => navigate('/')}
          className="mt-6 px-4 py-2 bg-gray-700 rounded-lg hover:bg-gray-600 transition-colors"
        >
          Back to Home
        </button>
      </div>
    );
  }

  const { council, responses: savedResponses, rankings: savedRankings } = data;

  // Merge everything by round and filter out invalid rounds
  const rounds = Array.from(new Set([
    ...(savedResponses?.map(r => r.round) || []),
    ...(Object.values(streamData.responses).map(r => r.round)),
    ...(council.syntheses?.map(s => s.round) || []),
    ...(Object.values(streamData.syntheses).map(s => s.round))
  ]))
    .filter(round => typeof round === 'number' && !isNaN(round) && round > 0)
    .sort((a, b) => a - b);

  const currentRound = rounds.length > 0 ? Math.max(...rounds) : 1;

  // Calculate per-member stats (excluding chairman)
  const allResponses = [
    ...(savedResponses || []),
    ...Object.values(streamData.responses)
  ];
  const allRankings = [
    ...(savedRankings || []),
    ...Object.values(streamData.rankings)
  ];

  const memberStats = council.members.map(member => {
    const memberResponses = allResponses.filter(r => r.member_id === member.id);
    const memberRankings = allRankings.filter(r => r.reviewer_id === member.id);

    let inputTokens = 0;
    let outputTokens = 0;
    let totalDuration = 0;

    memberResponses.forEach(r => {
      inputTokens += r.input_tokens || 0;
      outputTokens += r.output_tokens || 0;
      totalDuration += r.duration_ms || 0;
    });

    memberRankings.forEach(r => {
      inputTokens += r.input_tokens || 0;
      outputTokens += r.output_tokens || 0;
      totalDuration += r.duration_ms || 0;
    });

    return {
      member,
      inputTokens,
      outputTokens,
      totalDuration,
      hasData: inputTokens > 0 || outputTokens > 0
    };
  });

  return (
    <div className="max-w-6xl mx-auto py-4 px-4 space-y-12">
      {/* Header */}
      <div className="bg-brand-card shadow-xl rounded-xl p-8 border border-brand-border">
        <h1 className="text-xl font-bold text-white mb-3">{council.title || council.topic}</h1>
        <p className="text-gray-400 text-lg mb-6 whitespace-pre-wrap">{council.title ? council.topic : ''}</p>

        {/* Members with stats */}
        <div className="space-y-3 mb-6">
          {memberStats.map((stat, idx) => (
            <div key={stat.member.id} className="flex items-center gap-3 text-sm">
              <span className="w-6 h-6 rounded-full bg-brand-bg flex items-center justify-center text-xs">
                {['💭', '🧠', '🎯', '💡', '🔬'][idx % 5]}
              </span>
              <span className="text-[#d3c6aa] font-medium min-w-[180px]">{stat.member.name}</span>
              {stat.hasData ? (
                <span className="text-[#859289] font-mono text-xs">
                  ↑{formatTokens(stat.inputTokens)} ↓{formatTokens(stat.outputTokens)} • {formatDuration(stat.totalDuration)}
                </span>
              ) : (
                <span className="text-[#859289]/50 text-xs italic">waiting...</span>
              )}
            </div>
          ))}
        </div>

        {/* Chairman and directory info */}
        <div className="flex items-center gap-6 text-sm text-[#859289] pt-4 border-t border-brand-border">
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5 text-brand-secondary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            Chairman: {council.chairman.name}
          </span>
          {council.cwd && (
            <span className="flex items-center gap-2 font-mono text-xs bg-brand-bg px-2 py-1 rounded border border-brand-border">
              Dir: {council.cwd}
            </span>
          )}
        </div>
      </div>

      {/* Resume Status Indicator */}
      {council.status === 'in_progress' && rounds.length > 0 && (
        <div className="animate-fadeIn bg-gradient-to-r from-blue-500/10 via-brand-primary/10 to-blue-500/10 border-2 border-blue-500/30 rounded-xl p-5 shadow-lg">
          <div className="flex items-center gap-4">
            <div className="relative">
              <div className="w-3 h-3 bg-blue-500 rounded-full animate-ping absolute"></div>
              <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
            </div>
            <div className="flex-1">
              <div className="text-blue-400 font-bold text-lg">
                {rounds.length > 1 ? 'Re-convening Council' : 'Council Convening'}
              </div>
              <div className="text-sm text-[#859289]">
                Round {rounds.length > 0 ? Math.max(...rounds) : 1} in progress - members are {rounds.length > 1 ? 'responding to your directive' : 'analyzing the topic'}...
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

      {/* Synthesis Evolution Timeline */}
      {council.syntheses && council.syntheses.length > 1 && council.status === 'completed' && (
        <div className="animate-fadeIn bg-gradient-to-br from-brand-card via-brand-card to-brand-bg rounded-2xl border border-brand-border p-8 shadow-xl">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h3 className="text-lg font-bold text-[#d3c6aa]">Synthesis Evolution</h3>
              <p className="text-sm text-[#859289]">Track how the council's collective understanding evolved across {council.syntheses.length} rounds</p>
            </div>
          </div>

          <div className="space-y-3">
            {council.syntheses.map((synthesis, idx) => (
              <div key={idx} className="flex items-start gap-4 group">
                <div className="flex flex-col items-center">
                  <div className="w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm border-2 bg-green-500/20 text-green-500 border-green-500">
                    R{synthesis.round}
                  </div>
                  {idx < (council.syntheses?.length || 0) - 1 && (
                    <div className="w-0.5 h-full min-h-[40px] bg-brand-border mt-2"></div>
                  )}
                </div>

                <div className="flex-1 p-5 rounded-xl border-2 border-green-700/30 bg-green-900/10 transition-all group-hover:shadow-lg group-hover:bg-green-900/20">
                  <div className="flex items-center justify-between mb-3">
                    <div className="flex items-center gap-2">
                      <span className="text-lg">🏛️</span>
                      <span className="text-xs font-bold uppercase tracking-wider text-green-500">
                        Chairman's Synthesis
                      </span>
                    </div>
                    {idx === (council.syntheses?.length || 0) - 1 && (
                      <span className="text-xs bg-brand-bg px-2 py-1 rounded-full text-green-500 border border-green-500/30">
                        Latest
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-[#d3c6aa] line-clamp-3 group-hover:line-clamp-none transition-all">
                    {synthesis.content}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Rounds */}
      {rounds.map((round) => {
        const roundResponses = Array.from(new Map([
          ...(savedResponses?.filter(r => r.round === round) || []),
          ...Object.values(streamData.responses).filter(r => r.round === round)
        ].map(r => [r.id, r])).values());

        const roundRankings = Array.from(new Map([
          ...(savedRankings?.filter(r => r.round === round) || []),
          ...Object.values(streamData.rankings).filter(r => r.round === round)
        ].map(r => [r.id || `${r.round}-${r.reviewer_id}`, r])).values());

        const roundSynthesis = council.syntheses?.find(s => s.round === round) || streamData.syntheses[round];

        // Helper function to get member color
        const getMemberColor = (memberIndex: number) => {
          const colors: Array<'primary' | 'secondary' | 'blue' | 'purple' | 'orange'> = ['primary', 'secondary', 'blue', 'purple', 'orange'];
          return colors[memberIndex % colors.length];
        };

        // Helper to get member emoji
        const getMemberEmoji = (memberIndex: number) => {
          const emojis = ['💭', '🧠', '🎯', '💡', '🔬'];
          return emojis[memberIndex % emojis.length];
        };

        // Filter members who have responses/rankings to prevent empty stage boxes
        const membersWithResponses = council.members.filter(member =>
          roundResponses.find(r => r.member_id === member.id)
        );
        const membersWithRankings = council.members.filter(member =>
          roundRankings.find(r => r.reviewer_id === member.id)
        );

        const isStage1Complete = membersWithResponses.length === council.members.length;
        const isStage2Complete = membersWithRankings.length === council.members.length;
        const isLatestRound = round === currentRound;
        const isInProgress = council.status === 'in_progress' && isLatestRound;

        return (
          <RoundContainer
            key={round}
            roundNumber={round}
            stage={roundSynthesis ? '✓ Synthesized' : `${roundResponses.filter(r => r.member_id !== 'user').length}/${council.members.length} responses`}
          >
            {/* User message */}
            {roundResponses.filter(r => r.member_id === 'user').map(userMsg => (
              <Message.Root
                key={userMsg.id}
                role="user"
                name="You"
                timestamp={userMsg.created_at}
              >
                {userMsg.content}
              </Message.Root>
            ))}

            {/* Stage 1: Member Perspectives */}
            {(membersWithResponses.length > 0 || (isInProgress && !isStage1Complete)) && (
              <>
                <Message.Root role="system">
                  <span className="text-xs font-bold uppercase tracking-wider">Stage 1: Member Perspectives</span>
                </Message.Root>

                {membersWithResponses.map((member) => {
                  const memberIndex = council.members.findIndex(m => m.id === member.id);
                  const response = roundResponses.find(r => r.member_id === member.id);

                  return (
                    <Message.Root
                      key={response!.id}
                      role="agent"
                      name={member.name}
                      avatar={getMemberEmoji(memberIndex)}
                      agentColor={getMemberColor(memberIndex)}
                      timestamp={response!.created_at}
                      metadata={buildMetadata(response!)}
                    >
                      {response!.content}
                    </Message.Root>
                  );
                })}

                {isInProgress && !isStage1Complete && (
                  <div className="p-6 bg-brand-bg/30 rounded-xl border border-dashed border-brand-border animate-pulse flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="flex gap-1">
                        <div className="w-2 h-2 bg-brand-primary rounded-full animate-bounce [animation-delay:-0.3s]"></div>
                        <div className="w-2 h-2 bg-brand-primary rounded-full animate-bounce [animation-delay:-0.15s]"></div>
                        <div className="w-2 h-2 bg-brand-primary rounded-full animate-bounce"></div>
                      </div>
                      <div>
                        <div className="text-sm font-bold text-[#d3c6aa]">Deliberating...</div>
                        <div className="text-xs text-[#859289]">
                          {council.members.length - membersWithResponses.length} members are preparing their perspectives
                        </div>
                      </div>
                    </div>
                    <div className="text-xs font-mono text-[#859289] bg-brand-bg px-2 py-1 rounded">
                      STAGE 1 / 3
                    </div>
                  </div>
                )}
              </>
            )}

            {/* Stage 2: Peer Evaluations */}
            {(membersWithRankings.length > 0 || (isInProgress && isStage1Complete && !isStage2Complete)) && (
              <>
                <Message.Root role="system">
                  <span className="text-xs font-bold uppercase tracking-wider">Stage 2: Peer Evaluations</span>
                </Message.Root>

                {membersWithRankings.map((member) => {
                  const memberIndex = council.members.findIndex(m => m.id === member.id);
                  const ranking = roundRankings.find(r => r.reviewer_id === member.id);

                  return (
                    <Message.Root
                      key={`${ranking!.round}-${ranking!.reviewer_id}`}
                      role="agent"
                      name={`${member.name} (Evaluation)`}
                      avatar="📊"
                      agentColor={getMemberColor(memberIndex)}
                      timestamp={ranking!.created_at}
                      metadata={buildMetadata(ranking!)}
                    >
                      {ranking!.reasoning}
                    </Message.Root>
                  );
                })}

                {isInProgress && isStage1Complete && !isStage2Complete && (
                  <div className="p-6 bg-brand-bg/30 rounded-xl border border-dashed border-brand-border animate-pulse flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="flex gap-1">
                        <div className="w-2 h-2 bg-brand-secondary rounded-full animate-bounce [animation-delay:-0.3s]"></div>
                        <div className="w-2 h-2 bg-brand-secondary rounded-full animate-bounce [animation-delay:-0.15s]"></div>
                        <div className="w-2 h-2 bg-brand-secondary rounded-full animate-bounce"></div>
                      </div>
                      <div>
                        <div className="text-sm font-bold text-[#d3c6aa]">Evaluating Perspectives...</div>
                        <div className="text-xs text-[#859289]">
                          Members are reviewing and ranking each other's responses
                        </div>
                      </div>
                    </div>
                    <div className="text-xs font-mono text-[#859289] bg-brand-bg px-2 py-1 rounded">
                      STAGE 2 / 3
                    </div>
                  </div>
                )}
              </>
            )}

            {/* Stage 3: Synthesis */}
            {(roundSynthesis || (isInProgress && isStage2Complete)) && (
              <>
                <Message.Root role="system">
                  <span className="text-xs font-bold uppercase tracking-wider">Stage 3: Chairman's Synthesis</span>
                </Message.Root>
                {roundSynthesis ? (
                  <Message.Root
                    role="agent"
                    name={council.chairman.name}
                    avatar="🏛️"
                    agentColor="primary"
                    timestamp={roundSynthesis.created_at}
                    metadata={buildMetadata(roundSynthesis)}
                  >
                    {roundSynthesis.content}
                  </Message.Root>
                ) : (
                  <div className="p-6 bg-brand-bg/30 rounded-xl border border-dashed border-brand-border animate-pulse flex items-center justify-between">
                    <div className="flex items-center gap-4">
                      <div className="flex gap-1">
                        <div className="w-2 h-2 bg-brand-primary rounded-full animate-bounce [animation-delay:-0.3s]"></div>
                        <div className="w-2 h-2 bg-brand-primary rounded-full animate-bounce [animation-delay:-0.15s]"></div>
                        <div className="w-2 h-2 bg-brand-primary rounded-full animate-bounce"></div>
                      </div>
                      <div>
                        <div className="text-sm font-bold text-[#d3c6aa]">Finalizing Synthesis...</div>
                        <div className="text-xs text-[#859289]">
                          Chairman {council.chairman.name} is consolidating the council's findings
                        </div>
                      </div>
                    </div>
                    <div className="text-xs font-mono text-[#859289] bg-brand-bg px-2 py-1 rounded">
                      STAGE 3 / 3
                    </div>
                  </div>
                )}
              </>
            )}
          </RoundContainer>
        );
      })}

      {/* Empty state while starting */}
      {rounds.length === 0 && council.status === 'in_progress' && (
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
                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                />
              </svg>
            </div>
            <h3 className="text-3xl font-bold text-white mb-3">Council is convening...</h3>
            <p className="text-gray-400 text-lg max-w-md mx-auto">
              Council members are analyzing the topic and preparing their perspectives.
            </p>
          </div>
        </div>
      )}

      {/* Follow-up Input */}
      {council.status === 'completed' && (
        <div className="bg-brand-card rounded-2xl border border-brand-border p-8 shadow-2xl animate-fadeIn">
          <div className="flex items-center gap-3 mb-6">
            <div className="w-10 h-10 bg-brand-primary bg-opacity-20 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
              </svg>
            </div>
            <div>
              <h3 className="text-lg font-bold text-[#d3c6aa]">Guide the Council</h3>
              <p className="text-sm text-[#859289]">Add a follow-up directive or ask the members to re-evaluate based on new info.</p>
            </div>
          </div>

          <form onSubmit={handleFollowUp} className="space-y-4">
            <textarea
              value={followUp}
              onChange={(e) => setFollowUp(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="e.g., Now consider this from a historical perspective..."
              rows={3}
              className="w-full bg-brand-bg border-2 border-brand-border rounded-xl p-4 text-[#d3c6aa] focus:border-brand-primary outline-none transition-all"
            />
            <div className="flex justify-end">
              <button
                type="submit"
                disabled={!followUp.trim() || followUpMutation.isPending}
                className="px-6 py-3 bg-brand-primary hover:bg-[#b8cc95] text-[#2b3339] font-bold rounded-lg transition-all transform active:scale-95 disabled:opacity-50"
              >
                {followUpMutation.isPending ? 'Resuming...' : 'Re-convene Council'}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}

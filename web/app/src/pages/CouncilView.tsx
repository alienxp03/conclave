import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { api } from '../lib/api';
import type { CouncilResponse, CouncilRanking, CouncilSynthesis } from '../types';

export function CouncilView() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Active tabs state (mapped by round)
  const [activeResponseTabs, setActiveResponseTabs] = useState<Record<number, string>>({});
  const [activeRankingTabs, setActiveRankingTabs] = useState<Record<number, string>>({});
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

  const [completedStagesByRound, setCompletedStagesByRound] = useState<Record<number, number[]>>({});

  const followUpMutation = useMutation({
    mutationFn: (content: string) => api.addCouncilFollowUp(id!, content),
    onSuccess: () => {
      setFollowUp('');
      queryClient.invalidateQueries({ queryKey: ['council', id] });
    },
  });

  // Sync completed stages
  useEffect(() => {
    if (!data) return;

    const memberCount = data.council.members.length;
    const stagesByRound: Record<number, number[]> = {};

    // Get all rounds
    const rounds = Array.from(new Set([
      ...(data.responses?.map(r => r.round) || []),
      ...(Object.values(streamData.responses).map(r => r.round)),
      ...(data.council.syntheses?.map(s => s.round) || []),
      ...(Object.values(streamData.syntheses).map(s => s.round))
    ])).sort((a, b) => a - b);

    rounds.forEach(round => {
      const stages: number[] = [];
      const roundResponses = [
        ...(data.responses?.filter(r => r.round === round) || []),
        ...Object.values(streamData.responses).filter(r => r.round === round)
      ];
      
      // Stage 1 complete if all agents responded OR if stage 2 has started
      const agentResponses = roundResponses.filter(r => r.member_id !== 'user');
      if (agentResponses.length === memberCount) {
        stages.push(1);
      }

      const roundRankings = [
        ...(data.rankings?.filter(r => r.round === round) || []),
        ...Object.values(streamData.rankings).filter(r => r.round === round)
      ];
      if (roundRankings.length === memberCount) {
        if (!stages.includes(1)) stages.push(1);
        stages.push(2);
      }

      const hasSynthesis = data.council.syntheses?.some(s => s.round === round) || !!streamData.syntheses[round];
      if (hasSynthesis) {
        if (!stages.includes(1)) stages.push(1);
        if (!stages.includes(2)) stages.push(2);
        stages.push(3);
      }
      stagesByRound[round] = stages;

      // Set initial active tabs for this round if not set
      if (data.council.members.length > 0) {
        if (!activeResponseTabs[round]) {
          setActiveResponseTabs(prev => ({ ...prev, [round]: data.council.members[0].id }));
        }
        if (!activeRankingTabs[round]) {
          setActiveRankingTabs(prev => ({ ...prev, [round]: data.council.members[0].id }));
        }
      }
    });

    setCompletedStagesByRound(stagesByRound);
  }, [data, streamData]);

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
        const round = payload.round || Math.max(1, ...Object.values(prev.responses).map(r => r.round));
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

  // Merge everything by round
  const rounds = Array.from(new Set([
    ...(savedResponses?.map(r => r.round) || []),
    ...(Object.values(streamData.responses).map(r => r.round)),
    ...(council.syntheses?.map(s => s.round) || []),
    ...(Object.values(streamData.syntheses).map(s => s.round))
  ])).sort((a, b) => a - b);

  return (
    <div className="max-w-6xl mx-auto py-4 px-4 space-y-12">
      {/* Header */}
      <div className="bg-brand-card shadow-xl rounded-xl p-8 border border-brand-border">
        <h1 className="text-3xl font-bold text-white mb-3">{council.title || council.topic}</h1>
        <p className="text-gray-400 text-lg mb-6">{council.title ? council.topic : ''}</p>
        <div className="flex items-center gap-6 text-sm text-[#859289]">
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            {council.members.length} Members
          </span>
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
      {council.status === 'in_progress' && rounds.length > 1 && (
        <div className="animate-fadeIn bg-gradient-to-r from-blue-500/10 via-brand-primary/10 to-blue-500/10 border-2 border-blue-500/30 rounded-xl p-5 shadow-lg">
          <div className="flex items-center gap-4">
            <div className="relative">
              <div className="w-3 h-3 bg-blue-500 rounded-full animate-ping absolute"></div>
              <div className="w-3 h-3 bg-blue-500 rounded-full"></div>
            </div>
            <div className="flex-1">
              <div className="text-blue-400 font-bold text-lg">Re-convening Council</div>
              <div className="text-sm text-[#859289]">
                Round {Math.max(...rounds)} in progress - members are responding to your directive...
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
        const roundResponses = [
          ...(savedResponses?.filter(r => r.round === round) || []),
          ...Object.values(streamData.responses).filter(r => r.round === round)
        ];
        const roundRankings = [
          ...(savedRankings?.filter(r => r.round === round) || []),
          ...Object.values(streamData.rankings).filter(r => r.round === round)
        ];
        const roundSynthesis = council.syntheses?.find(s => s.round === round) || streamData.syntheses[round];
        const userDirective = roundResponses.find(r => r.member_id === 'user');
        const stages = completedStagesByRound[round] || [];

        const isStage1Active = council.status === 'in_progress' && !stages.includes(1);
        const isStage2Active = council.status === 'in_progress' && stages.includes(1) && !stages.includes(2);
        const isStage3Active = council.status === 'in_progress' && stages.includes(2) && !stages.includes(3);

        return (
          <div key={round} className="space-y-8 animate-fadeIn">
            <div className="flex items-center gap-4">
              <div className="h-px flex-1 bg-brand-border" />
              <div className="flex items-center gap-3">
                <span className="text-xs font-bold text-[#859289] uppercase tracking-widest bg-brand-bg px-4 py-1 rounded-full border border-brand-border">
                  Deliberation Round {round}
                </span>
                {roundSynthesis && (
                  <span className="text-xs px-2 py-1 rounded-full font-medium bg-green-500/20 text-green-500">
                    ✓ Synthesized
                  </span>
                )}
                <span className="text-xs text-[#859289] bg-brand-bg px-2 py-0.5 rounded-full">
                  {roundResponses.filter(r => r.member_id !== 'user').length}/{council.members.length} responses
                </span>
              </div>
              <div className="h-px flex-1 bg-brand-border" />
            </div>

            {userDirective && (
              <div className="bg-brand-primary/10 border-l-4 border-brand-primary p-6 rounded-r-xl shadow-inner">
                <div className="flex items-center gap-3 mb-2">
                  <span className="text-xs font-bold text-brand-primary uppercase tracking-wider">User Follow-up</span>
                </div>
                <p className="text-[#d3c6aa] text-lg italic italic">"{userDirective.content}"</p>
              </div>
            )}

            {/* Round Stages */}
            <div className="grid grid-cols-1 gap-12">
              {/* Stage 1: Responses */}
              <div className="space-y-6">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-brand-primary/20 text-brand-primary flex items-center justify-center font-bold border border-brand-primary/30">1</div>
                  <h2 className="text-xl font-bold text-white">Member Perspectives</h2>
                  {stages.includes(1) && <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                </div>

                <div className="bg-brand-card/30 rounded-xl border border-brand-border overflow-hidden shadow-lg">
                  <div className="flex flex-wrap border-b border-brand-border bg-brand-bg/20">
                    {council.members.map((member) => {
                      const isActive = activeResponseTabs[round] === member.id;
                      const response = roundResponses.find(r => r.member_id === member.id);
                      return (
                        <button
                          key={member.id}
                          onClick={() => setActiveResponseTabs(prev => ({ ...prev, [round]: member.id }))}
                          className={`px-6 py-3 text-sm font-bold transition-all border-b-2 ${isActive ? 'text-brand-primary border-brand-primary bg-brand-primary/5' : 'text-gray-500 border-transparent hover:text-gray-300'}`}
                        >
                          {member.name.match(/\(([^)]+)\)/)?.[1] || member.name}
                          {response && <span className="ml-2 inline-block w-1.5 h-1.5 rounded-full bg-green-500"></span>}
                          {!response && isStage1Active && <span className="ml-2 inline-block w-1.5 h-1.5 rounded-full bg-brand-primary animate-pulse"></span>}
                        </button>
                      );
                    })}
                  </div>
                  <div className="p-8 min-h-[200px]">
                    {activeResponseTabs[round] && (
                      <div className="prose prose-invert max-w-none">
                        {roundResponses.find(r => r.member_id === activeResponseTabs[round]) ? (
                          <ReactMarkdown remarkPlugins={[remarkGfm]}>{roundResponses.find(r => r.member_id === activeResponseTabs[round])!.content}</ReactMarkdown>
                        ) : (
                          <div className="text-gray-500 italic py-10 text-center">
                            {isStage1Active ? 'Thinking...' : 'Waiting for response...'}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Stage 2: Rankings */}
              <div className={`space-y-6 transition-opacity duration-500 ${stages.length >= 1 ? 'opacity-100' : 'opacity-30 pointer-events-none'}`}>
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-brand-accent/20 text-brand-accent flex items-center justify-center font-bold border border-brand-accent/30">2</div>
                  <h2 className="text-xl font-bold text-white">Peer Evaluations</h2>
                  {stages.includes(2) && <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                </div>

                <div className="bg-brand-card/30 rounded-xl border border-brand-border overflow-hidden shadow-lg">
                  <div className="flex flex-wrap border-b border-brand-border bg-brand-bg/20">
                    {council.members.map((member) => {
                      const isActive = activeRankingTabs[round] === member.id;
                      const ranking = roundRankings.find(r => r.reviewer_id === member.id);
                      return (
                        <button
                          key={member.id}
                          onClick={() => setActiveRankingTabs(prev => ({ ...prev, [round]: member.id }))}
                          className={`px-6 py-3 text-sm font-bold transition-all border-b-2 ${isActive ? 'text-brand-accent border-brand-accent bg-brand-accent/5' : 'text-gray-500 border-transparent hover:text-gray-300'}`}
                        >
                          {member.name.match(/\(([^)]+)\)/)?.[1] || member.name}
                          {ranking && <span className="ml-2 inline-block w-1.5 h-1.5 rounded-full bg-green-500"></span>}
                          {!ranking && isStage2Active && <span className="ml-2 inline-block w-1.5 h-1.5 rounded-full bg-brand-accent animate-pulse"></span>}
                        </button>
                      );
                    })}
                  </div>
                  <div className="p-8 min-h-[200px]">
                    {activeRankingTabs[round] && (
                      <div className="prose prose-invert max-w-none">
                        {roundRankings.find(r => r.reviewer_id === activeRankingTabs[round]) ? (
                          <ReactMarkdown remarkPlugins={[remarkGfm]}>{roundRankings.find(r => r.reviewer_id === activeRankingTabs[round])!.reasoning}</ReactMarkdown>
                        ) : (
                          <div className="text-gray-500 italic py-10 text-center">
                            {isStage2Active ? 'Evaluating...' : 'Waiting for evaluations...'}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Stage 3: Synthesis */}
              <div className={`space-y-6 transition-opacity duration-500 ${stages.length >= 2 ? 'opacity-100' : 'opacity-30 pointer-events-none'}`}>
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-green-500/20 text-green-500 flex items-center justify-center font-bold border border-green-500/30">3</div>
                  <h2 className="text-xl font-bold text-white">Consensus Synthesis</h2>
                  {stages.includes(3) && <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" /></svg>}
                </div>

                <div className="bg-green-900/10 rounded-xl border border-green-700/30 p-8 shadow-xl">
                  {roundSynthesis ? (
                    <div className="prose prose-invert max-w-none prose-strong:text-brand-primary">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>{roundSynthesis.content}</ReactMarkdown>
                    </div>
                  ) : (
                    <div className="text-center py-12 text-gray-500">
                      {isStage3Active ? (
                        <div className="flex flex-col items-center gap-3">
                          <div className="animate-spin h-6 w-6 border-2 border-gray-600 border-t-green-500 rounded-full"></div>
                          <span>Chairman is synthesizing...</span>
                        </div>
                      ) : 'Waiting...'}
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        );
      })}

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

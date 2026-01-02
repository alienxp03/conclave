import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { api } from '../lib/api';

export function CouncilView() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  // Active tabs
  const [activeResponseTab, setActiveResponseTab] = useState<string | null>(null);
  const [activeRankingTab, setActiveRankingTab] = useState<string | null>(null);

  // Streaming state
  const [streamData, setStreamData] = useState<{
    responses: Record<string, string>;
    rankings: Record<string, string>;
    synthesis: string | null;
  }>({
    responses: {},
    rankings: {},
    synthesis: null,
  });

  const { data, isLoading, error } = useQuery({
    queryKey: ['council', id],
    queryFn: () => api.getCouncil(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.council.status;
      return status === 'completed' || status === 'failed' ? false : 3000;
    },
  });

  const [completedStages, setCompletedStages] = useState<number[]>([]);

  // Sync completed stages
  useEffect(() => {
    if (!data) return;

    const memberCount = data.council.members.length;
    const stages: number[] = [];

    const responseCount = (data.responses?.length || 0);
    const streamResponseCount = Object.keys(streamData.responses).length;
    if (responseCount === memberCount || streamResponseCount === memberCount) {
      stages.push(1);
    }

    const rankingCount = (data.rankings?.length || 0);
    const streamRankingCount = Object.keys(streamData.rankings).length;
    if (rankingCount === memberCount || streamRankingCount === memberCount) {
      if (!stages.includes(1)) stages.push(1);
      stages.push(2);
    }

    if (data.council.status === 'completed' || !!data.council.synthesis || !!streamData.synthesis) {
      if (!stages.includes(1)) stages.push(1);
      if (!stages.includes(2)) stages.push(2);
      stages.push(3);
    }

    setCompletedStages(prev => {
      if (JSON.stringify(prev) === JSON.stringify(stages)) return prev;
      return stages;
    });

    // Set initial active tabs if not set
    if (data.council.members.length > 0) {
      if (!activeResponseTab) setActiveResponseTab(data.council.members[0].id);
      if (!activeRankingTab) setActiveRankingTab(data.council.members[0].id);
    }
  }, [data, streamData, activeResponseTab, activeRankingTab]);

  // Streaming effect
  useEffect(() => {
    if (!id || data?.council.status === 'completed') return;

    const eventSource = api.createCouncilStream(id);

    eventSource.onmessage = () => {};

    eventSource.addEventListener('response_collected', (e: any) => {
      const payload = JSON.parse(e.data);
      setStreamData(prev => ({
        ...prev,
        responses: { ...prev.responses, [payload.agent_id]: payload.content }
      }));
    });

    eventSource.addEventListener('ranking_collected', (e: any) => {
      const payload = JSON.parse(e.data);
      setStreamData(prev => ({
        ...prev,
        rankings: { ...prev.rankings, [payload.agent_id]: payload.content }
      }));
    });

    eventSource.addEventListener('synthesis_complete', (e: any) => {
      const payload = JSON.parse(e.data);
      setStreamData(prev => ({
        ...prev,
        synthesis: payload.synthesis
      }));
    });

    eventSource.addEventListener('stage_complete', (e: any) => {
      const payload = JSON.parse(e.data);
      setCompletedStages(prev => prev.includes(payload.stage) ? prev : [...prev, payload.stage]);
    });

    eventSource.addEventListener('complete', () => {
      queryClient.invalidateQueries({ queryKey: ['council', id] });
      setCompletedStages([1, 2, 3]);
      eventSource.close();
    });

    eventSource.addEventListener('error', (e) => {
      console.error('Stream error', e);
      eventSource.close();
    });

    return () => eventSource.close();
  }, [id, data?.council.status, queryClient]);

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

  // Merge stream with saved data
  const allResponses = { ...streamData.responses };
  savedResponses?.forEach(r => allResponses[r.member_id] = r.content);

  const allRankings = { ...streamData.rankings };
  savedRankings?.forEach(r => allRankings[r.reviewer_id] = r.reasoning);

  const finalSynthesis = council.synthesis || streamData.synthesis;

  const isStage1Active = council.status === 'in_progress' && !completedStages.includes(1);
  const isStage2Active = council.status === 'in_progress' && completedStages.includes(1) && !completedStages.includes(2);
  const isStage3Active = council.status === 'in_progress' && completedStages.includes(2) && !completedStages.includes(3);

  return (
    <div className="max-w-6xl mx-auto py-4 px-4">
      {/* Header */}
      <div className="mb-8">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-white mb-3">{council.title || council.topic}</h1>
          <p className="text-gray-400 mb-4">{council.title ? council.topic : ''}</p>
          <div className="flex items-center gap-4 text-sm text-gray-400">
            <span className="flex items-center gap-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              {council.members.length} Members
            </span>
            <span>•</span>
            <span>Chairman: {council.chairman.name}</span>
          </div>
          {council.cwd && (
            <div className="mt-4 flex items-center text-[#d3c6aa] text-xs font-mono bg-brand-bg bg-opacity-40 px-3 py-1.5 rounded border border-brand-border w-fit shadow-inner">
              <svg className="w-3.5 h-3.5 mr-1.5 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
              <span className="opacity-60 mr-1 text-[#859289]">Session dir:</span> {council.cwd}
            </div>
          )}
        </div>
      </div>

      {/* Stage 1 */}
      <div className="mb-16 pb-12 border-b-2 border-brand-border">
        <div className="flex items-center gap-3 mb-8">
          <div className="flex items-center justify-center w-10 h-10 rounded-full bg-brand-primary text-white font-bold text-base shadow-lg">
            1
          </div>
          <h2 className="text-2xl font-bold text-white">Individual Responses</h2>
          {completedStages.includes(1) && (
            <svg className="w-6 h-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
            </svg>
          )}
        </div>

        {/* Tabs for responses */}
        <div className="bg-brand-card/50 rounded-xl border-2 border-brand-border shadow-lg overflow-hidden">
          <div className="flex flex-wrap border-b border-brand-border bg-brand-bg/50">
            {council.members.map((member) => {
              const perspectiveMatch = member.name.match(/\(([^)]+)\)/);
              const perspective = perspectiveMatch ? perspectiveMatch[1] : member.name;
              const isActive = activeResponseTab === member.id;
              const hasResponse = !!allResponses[member.id];
              const modelLabel = member.model ? `${member.provider}/${member.model}` : member.provider;

              return (
                <button
                  key={member.id}
                  onClick={() => setActiveResponseTab(member.id)}
                  className={`px-4 py-2 text-sm font-bold transition-all border-b-2 ${
                    isActive
                      ? 'text-brand-primary border-brand-primary bg-brand-primary/5'
                      : 'text-gray-500 border-transparent hover:text-gray-300 hover:bg-white/5'
                  }`}
                >
                  <div className="flex flex-col items-center gap-0.5">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-mono tracking-tight">{modelLabel}</span>
                      {hasResponse && (
                        <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                      )}
                      {!hasResponse && isStage1Active && (
                        <div className="w-1.5 h-1.5 rounded-full bg-brand-primary animate-pulse"></div>
                      )}
                    </div>
                    <span className="text-xs opacity-60 font-bold tracking-wide">{perspective}</span>
                  </div>
                </button>
              );
            })}
          </div>
          <div className="p-8 min-h-[300px]">
            {activeResponseTab && (
              <div key={activeResponseTab} className="animate-fadeIn">
                {allResponses[activeResponseTab] ? (
                  <div className="prose prose-invert max-w-none">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {allResponses[activeResponseTab]}
                    </ReactMarkdown>
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center py-20 text-gray-500 gap-4">
                    {isStage1Active ? (
                      <>
                        <div className="animate-spin h-8 w-8 border-4 border-gray-700 border-t-blue-500 rounded-full"></div>
                        <p className="text-lg">Thinking...</p>
                      </>
                    ) : council.status === 'failed' ? (
                      <p className="text-lg text-red-400">Failed to get response.</p>
                    ) : (
                      <p className="text-lg">Waiting for response...</p>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Stage 2 */}
      <div className="mb-16 pb-12 border-b-2 border-brand-border">
        <div className="flex items-center gap-3 mb-8">
          <div className="flex items-center justify-center w-10 h-10 rounded-full bg-brand-accent text-white font-bold text-base shadow-lg">
            2
          </div>
          <h2 className="text-2xl font-bold text-white">Peer Rankings</h2>
          {completedStages.includes(2) && (
            <svg className="w-6 h-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
            </svg>
          )}
        </div>

        <div className="bg-brand-card/50 rounded-xl border-2 border-brand-border shadow-lg overflow-hidden">
          <div className="flex flex-wrap border-b border-brand-border bg-brand-bg/50">
            {council.members.map((member) => {
              const perspectiveMatch = member.name.match(/\(([^)]+)\)/);
              const perspective = perspectiveMatch ? perspectiveMatch[1] : member.name;
              const isActive = activeRankingTab === member.id;
              const hasRanking = !!allRankings[member.id];
              const modelLabel = member.model ? `${member.provider}/${member.model}` : member.provider;

              return (
                <button
                  key={member.id}
                  onClick={() => setActiveRankingTab(member.id)}
                  className={`px-4 py-2 text-sm font-bold transition-all border-b-2 ${
                    isActive
                      ? 'text-brand-accent border-brand-accent bg-brand-accent/5'
                      : 'text-gray-500 border-transparent hover:text-gray-300 hover:bg-white/5'
                  }`}
                >
                  <div className="flex flex-col items-center gap-0.5">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-mono tracking-tight">{modelLabel}</span>
                      {hasRanking && (
                        <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                      )}
                      {!hasRanking && isStage2Active && (
                        <div className="w-1.5 h-1.5 rounded-full bg-brand-accent animate-pulse"></div>
                      )}
                    </div>
                    <span className="text-xs opacity-60 font-bold tracking-wide">{perspective}</span>
                  </div>
                </button>
              );
            })}
          </div>
          <div className="p-8 min-h-[300px]">
            {activeRankingTab && (
              <div key={activeRankingTab} className="animate-fadeIn">
                {allRankings[activeRankingTab] ? (
                  <div className="prose prose-invert max-w-none">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {allRankings[activeRankingTab]}
                    </ReactMarkdown>
                  </div>
                ) : (
                  <div className="flex flex-col items-center justify-center py-20 text-gray-500 gap-4">
                    {isStage2Active ? (
                      <>
                        <div className="animate-spin h-8 w-8 border-4 border-gray-700 border-t-purple-500 rounded-full"></div>
                        <p className="text-lg">Reviewing responses...</p>
                      </>
                    ) : council.status === 'failed' ? (
                      <p className="text-lg text-red-400">Failed to get ranking.</p>
                    ) : (
                      <p className="text-lg">Waiting for ranking...</p>
                    )}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Stage 3 */}
      <div className="mb-12">
        <div className="flex items-center gap-3 mb-8">
          <div className="flex items-center justify-center w-10 h-10 rounded-full bg-green-500 text-white font-bold text-base shadow-lg">
            3
          </div>
          <h2 className="text-2xl font-bold text-white">Final Council Answer</h2>
          {completedStages.includes(3) && (
            <svg className="w-6 h-6 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
            </svg>
          )}
        </div>

        <div className="text-sm text-gray-400 mb-6">
          <strong>Chairman:</strong> {council.chairman.name}
        </div>

         <div className="bg-green-900/20 rounded-xl border-2 border-green-700/40 p-8 shadow-xl">
           {finalSynthesis ? (
             <div className="prose prose-invert max-w-none">
               <ReactMarkdown remarkPlugins={[remarkGfm]}>
                 {finalSynthesis}
               </ReactMarkdown>
             </div>
           ) : (
            <div className="text-center py-12 text-gray-500">
              {isStage3Active ? (
                <div className="flex flex-col items-center gap-3">
                  <div className="animate-spin h-6 w-6 border-2 border-gray-600 border-t-green-500 rounded-full"></div>
                  <span>Chairman is deliberating...</span>
                </div>
              ) : (
                <span>Waiting...</span>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

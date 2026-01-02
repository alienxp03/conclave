import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';

export function CouncilView() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState<'responses' | 'rankings' | 'synthesis'>('responses');
  const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);

  // Streaming state
  const [streamData, setStreamData] = useState<{
    responses: Record<string, string>; // memberId -> content
    rankings: Record<string, string>;  // reviewerId -> content
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

  // Initialize selected agent
  useEffect(() => {
    if (data?.council.members.length && !selectedAgentId) {
      setSelectedAgentId(data.council.members[0].id);
    }
  }, [data, selectedAgentId]);

  // Streaming effect
  useEffect(() => {
    if (!id || data?.council.status === 'completed') return;

    const eventSource = api.createCouncilStream(id);

    eventSource.onmessage = () => {
      // Keep connection alive
    };

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

    eventSource.addEventListener('complete', () => {
      queryClient.invalidateQueries({ queryKey: ['council', id] });
      eventSource.close();
    });

    eventSource.addEventListener('error', (e) => {
      console.error('Stream error', e);
      eventSource.close();
    });

    return () => {
      eventSource.close();
    };
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

  // Merge stream data with saved data
  const allResponses = { ...streamData.responses };
  savedResponses?.forEach(r => allResponses[r.member_id] = r.content);

  const allRankings = { ...streamData.rankings };
  savedRankings?.forEach(r => allRankings[r.reviewer_id] = r.reasoning);

  const finalSynthesis = council.synthesis || streamData.synthesis;

  // Determine available tabs
  const hasRankings = Object.keys(allRankings).length > 0;
  const hasSynthesis = !!finalSynthesis;

  // Auto-switch tabs based on progress
  useEffect(() => {
    if (hasSynthesis && activeTab !== 'synthesis') {
      setActiveTab('synthesis');
    } else if (hasRankings && activeTab === 'responses' && !hasSynthesis) {
        // Stay on responses or switch? Maybe stay unless user clicks.
        // Actually, let's switch to show progress if we are just arriving
    }
  }, [hasSynthesis]);


  return (
    <div className="space-y-8 animate-fadeIn">
      {/* Header */}
      <div className="border-b border-gray-700 pb-6">
        <div className="flex justify-between items-start mb-4">
          <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-900 text-blue-200 border border-blue-700">
             {council.status === 'completed' ? 'Completed' : 'In Progress'}
          </span>
          <span className="text-sm text-gray-400">
             {new Date(council.created_at).toLocaleString()}
          </span>
        </div>
        <h1 className="text-3xl font-bold text-white mb-2 leading-tight">
          {council.topic}
        </h1>
        <div className="flex items-center gap-2 text-gray-400">
           <span>{council.members.length} Members</span>
           <span>•</span>
           <span>Chairman: {council.chairman.name}</span>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 bg-gray-800 p-1 rounded-xl">
        <button
          onClick={() => setActiveTab('responses')}
          className={`flex-1 py-2.5 px-4 rounded-lg text-sm font-medium transition-all duration-200 ${
            activeTab === 'responses'
              ? 'bg-blue-600 text-white shadow-lg'
              : 'text-gray-400 hover:text-white hover:bg-gray-700'
          }`}
        >
          Stage 1: First Opinions
        </button>
        <button
          onClick={() => setActiveTab('rankings')}
          disabled={!hasRankings && council.status !== 'completed'}
          className={`flex-1 py-2.5 px-4 rounded-lg text-sm font-medium transition-all duration-200 ${
            activeTab === 'rankings'
              ? 'bg-purple-600 text-white shadow-lg'
              : 'text-gray-400 hover:text-white hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed'
          }`}
        >
          Stage 2: Peer Reviews
        </button>
        <button
          onClick={() => setActiveTab('synthesis')}
          disabled={!hasSynthesis && council.status !== 'completed'}
          className={`flex-1 py-2.5 px-4 rounded-lg text-sm font-medium transition-all duration-200 ${
            activeTab === 'synthesis'
              ? 'bg-green-600 text-white shadow-lg'
              : 'text-gray-400 hover:text-white hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed'
          }`}
        >
          Stage 3: Final Synthesis
        </button>
      </div>

      {/* Content */}
      <div className="min-h-[500px]">
        
        {/* Stage 1 & 2: Split View (List + Content) */}
        {(activeTab === 'responses' || activeTab === 'rankings') && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            {/* Sidebar */}
            <div className="space-y-2">
              {council.members.map((member) => {
                const hasData = activeTab === 'responses' 
                    ? !!allResponses[member.id] 
                    : !!allRankings[member.id];
                
                return (
                  <button
                    key={member.id}
                    onClick={() => setSelectedAgentId(member.id)}
                    className={`w-full text-left p-3 rounded-lg border transition-all duration-200 ${
                      selectedAgentId === member.id
                        ? 'bg-gray-800 border-blue-500 ring-1 ring-blue-500 text-white'
                        : 'bg-gray-900 border-gray-700 text-gray-400 hover:bg-gray-800 hover:text-white'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                        <span className="font-semibold text-sm">{member.name.split('(')[0]}</span>
                        {hasData && <span className="text-green-400">✓</span>}
                    </div>
                    <div className="text-xs opacity-75 truncate">
                        {member.name.split('(')[1]?.replace(')', '') || member.persona}
                    </div>
                  </button>
                );
              })}
            </div>

            {/* Main Content Area */}
            <div className="md:col-span-3 bg-gray-800 rounded-xl border border-gray-700 p-6 shadow-xl">
              {selectedAgentId && (
                <>
                  <div className="flex items-center gap-3 mb-6 border-b border-gray-700 pb-4">
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold">
                      {council.members.find(m => m.id === selectedAgentId)?.name.charAt(0)}
                    </div>
                    <div>
                      <h3 className="text-xl font-bold text-white">
                        {council.members.find(m => m.id === selectedAgentId)?.name}
                      </h3>
                      <p className="text-sm text-gray-400">
                        {activeTab === 'responses' ? 'Initial Opinion' : 'Peer Review & Ranking'}
                      </p>
                    </div>
                  </div>
                  
                  <div className="prose prose-invert max-w-none">
                    {activeTab === 'responses' ? (
                       allResponses[selectedAgentId] ? (
                         <div className="whitespace-pre-wrap">{allResponses[selectedAgentId]}</div>
                       ) : (
                         <div className="flex flex-col items-center justify-center py-12 text-gray-500">
                            <div className="animate-pulse mb-2">Thinking...</div>
                         </div>
                       )
                    ) : (
                       allRankings[selectedAgentId] ? (
                         <div className="whitespace-pre-wrap">{allRankings[selectedAgentId]}</div>
                       ) : (
                         <div className="flex flex-col items-center justify-center py-12 text-gray-500">
                            {council.status === 'in_progress' ? 'Waiting for responses...' : 'No review provided'}
                         </div>
                       )
                    )}
                  </div>
                </>
              )}
            </div>
          </div>
        )}

        {/* Stage 3: Synthesis */}
        {activeTab === 'synthesis' && (
          <div className="bg-gray-800 rounded-xl border-2 border-green-500/30 p-8 shadow-2xl animate-fadeIn">
             <div className="flex items-center gap-4 mb-8 border-b border-gray-700 pb-6">
                <div className="w-16 h-16 rounded-full bg-gradient-to-r from-green-500 to-emerald-700 flex items-center justify-center text-white text-2xl shadow-lg">
                  ⚖️
                </div>
                <div>
                  <h2 className="text-3xl font-bold text-white">Final Verdict</h2>
                  <p className="text-green-400">Synthesized by Chairman {council.chairman.name}</p>
                </div>
             </div>

             <div className="prose prose-lg prose-invert max-w-none">
                {finalSynthesis ? (
                    <div className="whitespace-pre-wrap leading-relaxed">{finalSynthesis}</div>
                ) : (
                    <div className="flex flex-col items-center justify-center py-20 text-gray-500">
                        <div className="animate-spin h-8 w-8 border-2 border-green-500 border-t-transparent rounded-full mb-4"></div>
                        <p>The Chairman is deliberating...</p>
                    </div>
                )}
             </div>
          </div>
        )}

      </div>
    </div>
  );
}

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { CreateDebateRequest } from '../types';

const EXAMPLE_TOPICS = [
  'Is social media more harmful than beneficial?',
  'Should universal basic income be implemented?',
  'Is climate change the most urgent global challenge?',
];

export function NewDebate() {
  const navigate = useNavigate();
  const [topic, setTopic] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAdvanced, setShowAdvanced] = useState(false);

  // Form state
  const [config, setConfig] = useState<Partial<CreateDebateRequest>>({
    mode: 'automatic',
    max_turns: 5,
    auto_run: true,
  });

  const { data: providers } = useQuery({
    queryKey: ['providers'],
    queryFn: () => api.getProviders(),
  });

  const { data: personas } = useQuery({
    queryKey: ['personas'],
    queryFn: () => api.getPersonas(),
  });

  const { data: styles } = useQuery({
    queryKey: ['styles'],
    queryFn: () => api.getStyles(),
  });

  const isReady = !!(providers?.length && personas?.length && styles?.length);

  // Set defaults when data loads
  useEffect(() => {
    if (providers?.length && personas?.length && styles?.length) {
      const availableProviders = providers.filter(p => p.available && p.name !== 'mock');
      const defaultProvider = availableProviders.length > 0 ? availableProviders[0].name : '';
      
      setConfig(prev => ({
        ...prev,
        agent_a_provider: prev.agent_a_provider || defaultProvider,
        agent_b_provider: prev.agent_b_provider || defaultProvider,
        agent_a_persona: prev.agent_a_persona || personas[0].id,
        agent_b_persona: prev.agent_b_persona || personas[1]?.id || personas[0].id,
        style: prev.style || styles[0].id,
      }));
    }
  }, [providers, personas, styles]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!topic.trim()) return;
    if (!isReady) {
      setError('System is initializing, please wait...');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // Final check for empty provider
      if (!config.agent_a_provider || !config.agent_b_provider) {
        throw new Error('Please select a provider for both agents');
      }

      const request: CreateDebateRequest = {
        topic: topic.trim(),
        agent_a_provider: config.agent_a_provider,
        agent_a_model: config.agent_a_model || '',
        agent_a_persona: config.agent_a_persona || '',
        agent_b_provider: config.agent_b_provider,
        agent_b_model: config.agent_b_model || '',
        agent_b_persona: config.agent_b_persona || '',
        style: config.style || '',
        mode: config.mode || 'automatic',
        max_turns: config.max_turns || 5,
        auto_run: config.auto_run,
      };

      const debate = await api.createDebate(request);
      navigate(`/debates/${debate.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create debate');
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleSubmit(e as any);
    }
  };

  return (
    <div className="min-h-[70vh] flex flex-col items-center justify-center py-12">
      <div className="max-w-4xl w-full mx-auto px-4">
        <div className="mb-12 text-center">
          <div className="mb-6">
            <span className="text-6xl">🎭</span>
          </div>
          <h1 className="text-5xl font-bold text-white mb-4 leading-tight">
            What should we debate?
          </h1>
          <p className="text-xl text-gray-400">
            Enter a question and watch AI agents discuss different perspectives
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-8">
          {error && (
            <div className="animate-fadeIn bg-red-900 bg-opacity-50 border-2 border-red-500 text-red-200 px-6 py-4 rounded-xl shadow-lg">
              <div className="flex items-start">
                <svg
                  className="w-6 h-6 mr-3 flex-shrink-0"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <div>
                  <h3 className="font-semibold mb-1">Error</h3>
                  <p className="text-sm">{error}</p>
                </div>
              </div>
            </div>
          )}

          <div>
            <textarea
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
              required
              autoFocus
              placeholder="e.g., Should AI replace human decision-making in healthcare?"
              className="block w-full rounded-xl bg-gray-800 border-2 border-gray-600 text-white placeholder-gray-500 focus:border-blue-500 focus:ring-4 focus:ring-blue-500 focus:ring-opacity-30 text-xl p-6 transition-all duration-200 shadow-lg hover:shadow-2xl outline-none"
            />
          </div>

          <div className="text-center">
             <button
              type="button"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="text-sm text-gray-400 hover:text-white underline decoration-dotted underline-offset-4"
            >
              {showAdvanced ? 'Hide Advanced Options' : 'Show Advanced Options'}
            </button>
          </div>

          {showAdvanced && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 bg-gray-800 p-6 rounded-xl border border-gray-700 animate-fadeIn">
              <div className="space-y-4">
                <h3 className="text-lg font-medium text-white border-b border-gray-700 pb-2">Agent A</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Provider</label>
                  <select
                    value={config.agent_a_provider}
                    onChange={(e) => setConfig({ ...config, agent_a_provider: e.target.value })}
                    className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500"
                  >
                    {providers?.filter(p => p.name !== 'mock').map((p) => (
                      <option key={p.name} value={p.name} disabled={!p.available}>
                        {p.display_name} {!p.available && '(Unavailable)'}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Persona</label>
                  <select
                    value={config.agent_a_persona}
                    onChange={(e) => setConfig({ ...config, agent_a_persona: e.target.value })}
                    className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500"
                  >
                    {personas?.map((p) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="space-y-4">
                <h3 className="text-lg font-medium text-white border-b border-gray-700 pb-2">Agent B</h3>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Provider</label>
                  <select
                    value={config.agent_b_provider}
                    onChange={(e) => setConfig({ ...config, agent_b_provider: e.target.value })}
                    className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500"
                  >
                    {providers?.filter(p => p.name !== 'mock').map((p) => (
                      <option key={p.name} value={p.name} disabled={!p.available}>
                        {p.display_name} {!p.available && '(Unavailable)'}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-400 mb-1">Persona</label>
                  <select
                    value={config.agent_b_persona}
                    onChange={(e) => setConfig({ ...config, agent_b_persona: e.target.value })}
                    className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500"
                  >
                    {personas?.map((p) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="md:col-span-2 space-y-4 pt-4 border-t border-gray-700">
                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-1">Debate Style</label>
                        <select
                            value={config.style}
                            onChange={(e) => setConfig({ ...config, style: e.target.value })}
                            className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500"
                        >
                            {styles?.map((s) => (
                            <option key={s.id} value={s.id}>{s.name}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-1">Max Turns (per agent)</label>
                        <input
                            type="number"
                            min="1"
                            max="20"
                            value={config.max_turns}
                            onChange={(e) => setConfig({ ...config, max_turns: parseInt(e.target.value) })}
                            className="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500"
                        />
                    </div>
                 </div>
              </div>
            </div>
          )}

          <div className="flex flex-col items-center space-y-4">
             <p className="text-sm text-gray-500">
              Press Ctrl+Enter (⌘+Enter on Mac) to submit
            </p>
            <button
              type="submit"
              disabled={loading || !topic.trim() || !isReady}
              className="group inline-flex items-center px-12 py-5 border border-transparent rounded-xl text-xl font-semibold text-white bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-500 focus:ring-opacity-50 shadow-2xl transform transition-all duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
            >
              {loading ? (
                <>
                  <svg
                    className="animate-spin -ml-1 mr-3 h-6 w-6 text-white"
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
                  Creating...
                </>
              ) : (
                <>
                  <span className="group-hover:mr-1 transition-all duration-200">
                    Start Debate
                  </span>
                  <svg
                    className="ml-2 w-6 h-6 group-hover:translate-x-1 transition-transform duration-200"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M13 7l5 5m0 0l-5 5m5-5H6"
                    />
                  </svg>
                </>
              )}
            </button>
          </div>

          <div className="mt-8 text-center">
            <p className="text-sm text-gray-500 mb-3">Or try one of these:</p>
            <div className="flex flex-wrap justify-center gap-2">
              {EXAMPLE_TOPICS.map((example) => (
                <button
                  key={example}
                  type="button"
                  onClick={() => setTopic(example)}
                  className="px-4 py-2 bg-gray-800 border border-gray-700 rounded-lg text-sm text-gray-300 hover:bg-gray-700 hover:border-gray-600 transition-colors"
                >
                  {example.split(' ').slice(0, 3).join(' ')}...
                </button>
              ))}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
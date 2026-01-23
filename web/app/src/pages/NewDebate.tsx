import { useState, useEffect, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { CreateDebateRequest } from '../types';
import { ProviderHealthDashboard } from '../components/ProviderHealthDashboard';

const EXAMPLE_TOPICS = [
  'Is social media more harmful than beneficial?',
  'Should universal basic income be implemented?',
  'Is climate change the most urgent global challenge?',
];

export function NewDebate() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [topic, setTopic] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [projectId, setProjectId] = useState('');

  // Form state
  const [config, setConfig] = useState<Partial<CreateDebateRequest>>({
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

  const { data: systemInfo } = useQuery({
    queryKey: ['systemInfo'],
    queryFn: () => api.getSystemInfo(),
  });

  const { data: projects } = useQuery({
    queryKey: ['projects', 'all'],
    queryFn: () => api.getProjects(100, 0),
  });

  const isReady = !!(providers?.length && personas?.length && styles?.length);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  }, [topic]);

  // Set defaults when data loads
  useEffect(() => {
    if (providers?.length && personas?.length && styles?.length) {
      const availableProviders = providers.filter(p => p.available && p.name !== 'mock');
      const defaultProv = availableProviders.length > 0 ? availableProviders[0] : null;
      const defaultProviderName = defaultProv?.name || '';
      const defaultModelName = defaultProv?.default_model || '';
      
      setConfig(prev => ({
        ...prev,
        agent_a_provider: prev.agent_a_provider || defaultProviderName,
        agent_a_model: prev.agent_a_model || defaultModelName,
        agent_b_provider: prev.agent_b_provider || defaultProviderName,
        agent_b_model: prev.agent_b_model || defaultModelName,
        agent_a_persona: prev.agent_a_persona || personas[0].id,
        agent_b_persona: prev.agent_b_persona || personas[1]?.id || personas[0].id,
        style: prev.style || styles[0].id,
      }));
    }
  }, [providers, personas, styles]);

  useEffect(() => {
    const projectFromQuery = searchParams.get('project');
    if (projectFromQuery) {
      setProjectId(projectFromQuery);
    }
  }, [searchParams]);

  const handleProviderChange = (agent: 'a' | 'b', providerName: string) => {
    const provider = providers?.find(p => p.name === providerName);
    const defaultModel = provider?.default_model || '';
    
    if (agent === 'a') {
      setConfig({ ...config, agent_a_provider: providerName, agent_a_model: defaultModel });
    } else {
      setConfig({ ...config, agent_b_provider: providerName, agent_b_model: defaultModel });
    }
  };

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
        project_id: projectId || undefined,
        agent_a_provider: config.agent_a_provider,
        agent_a_model: config.agent_a_model || '',
        agent_a_persona: config.agent_a_persona || '',
        agent_b_provider: config.agent_b_provider,
        agent_b_model: config.agent_b_model || '',
        agent_b_persona: config.agent_b_persona || '',
        style: config.style || '',
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
    <div className="min-h-[70vh] flex flex-col items-start justify-center py-6 md:py-12">
      <div className="max-w-4xl w-full px-2 md:px-4">
        <div className="mb-8 md:mb-12 text-left">
          <div className="mb-4 md:mb-6 flex justify-start">
            <span className="text-brand-primary">
              <svg className="w-12 h-12 md:w-16 md:h-16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 22C17.5228 22 22 17.5228 22 12C22 6.47715 17.5228 2 12 2C6.47715 2 2 6.47715 2 12C2 13.5997 2.37562 15.1116 3.04348 16.4522L2 22L7.54777 20.9565C8.88837 21.6244 10.4003 22 12 22Z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M8 12H8.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M12 12H12.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                <path d="M16 12H16.01" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
              </svg>
            </span>
          </div>
          <h1 className="text-2xl md:text-4xl font-bold text-white mb-3 md:mb-4 leading-tight">
            What should we debate?
          </h1>
          <p className="text-base md:text-lg text-gray-400">
            Enter a question and watch AI agents discuss different perspectives
          </p>
          {systemInfo?.cwd && (
            <div className="mt-4 md:mt-6 flex items-center text-[#d3c6aa] text-[10px] md:text-xs font-mono bg-brand-bg bg-opacity-40 px-2 md:px-3 py-1 md:py-1.5 rounded border border-brand-border w-fit shadow-inner mx-auto lg:mx-0">
              <svg className="w-3 h-3 md:w-3.5 md:h-3.5 mr-1.5 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
              <span className="opacity-60 mr-1 text-[#859289]">Running in:</span> {systemInfo.cwd}
            </div>
          )}
        </div>

        <form onSubmit={handleSubmit} className="space-y-6 md:space-y-8">
          {error && (
            <div className="animate-fadeIn bg-red-900 bg-opacity-50 border-2 border-red-500 text-red-200 px-4 md:px-6 py-3 md:py-4 rounded-xl shadow-lg">
              <div className="flex items-start">
                <svg
                  className="w-5 h-5 md:w-6 md:h-6 mr-3 flex-shrink-0"
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
                  <h3 className="font-semibold text-sm md:text-base mb-1">Error</h3>
                  <p className="text-xs md:text-sm">{error}</p>
                </div>
              </div>
            </div>
          )}

          <ProviderHealthDashboard />

          <div>
            <textarea
              ref={textareaRef}
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
              required
              autoFocus
              placeholder="e.g., Should AI replace human decision-making in healthcare?"
              className="block w-full rounded-xl bg-brand-card border-2 border-brand-border text-[#d3c6aa] placeholder-[#5c6a72] focus:border-brand-primary focus:ring-4 focus:ring-brand-primary focus:ring-opacity-20 text-sm md:text-base p-4 md:p-6 transition-all duration-200 shadow-lg hover:shadow-2xl outline-none"
            />
          </div>

          <div>
            <label className="text-xs uppercase tracking-widest text-[#859289] font-semibold">Project</label>
            <select
              value={projectId}
              onChange={(e) => setProjectId(e.target.value)}
              className="mt-2 w-full bg-brand-bg border border-brand-border rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-brand-primary"
            >
              <option value="">No project</option>
              {(projects || []).map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name}
                </option>
              ))}
            </select>
          </div>

          <div className="text-center">
             <button
              type="button"
              onClick={() => setShowAdvanced(!showAdvanced)}
              className="text-xs md:text-sm text-[#859289] hover:text-brand-primary underline decoration-dotted underline-offset-4"
            >
              {showAdvanced ? 'Hide Advanced Options' : 'Show Advanced Options'}
            </button>
          </div>

          {showAdvanced && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6 bg-brand-card p-4 md:p-6 rounded-xl border border-brand-border animate-fadeIn">
              <div className="space-y-3 md:space-y-4">
                <h3 className="text-base md:text-lg font-medium text-[#d3c6aa] border-b border-brand-border pb-2">Agent A</h3>
                <div>
                  <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Provider</label>
                  <select
                    value={config.agent_a_provider}
                    onChange={(e) => handleProviderChange('a', e.target.value)}
                    className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                  >
                    {providers?.filter(p => p.available && p.name !== 'mock').map((p) => (
                      <option key={p.name} value={p.name}>{p.display_name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Model</label>
                  <select
                    value={config.agent_a_model}
                    onChange={(e) => setConfig({ ...config, agent_a_model: e.target.value })}
                    className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                  >
                    {providers?.find(p => p.name === config.agent_a_provider)?.models.map((m) => (
                      <option key={m} value={m}>{m}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Persona</label>
                  <select
                    value={config.agent_a_persona}
                    onChange={(e) => setConfig({ ...config, agent_a_persona: e.target.value })}
                    className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                  >
                    {personas?.map((p) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="space-y-3 md:space-y-4">
                <h3 className="text-base md:text-lg font-medium text-[#d3c6aa] border-b border-brand-border pb-2">Agent B</h3>
                <div>
                  <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Provider</label>
                  <select
                    value={config.agent_b_provider}
                    onChange={(e) => handleProviderChange('b', e.target.value)}
                    className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                  >
                    {providers?.filter(p => p.available && p.name !== 'mock').map((p) => (
                      <option key={p.name} value={p.name}>{p.display_name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Model</label>
                  <select
                    value={config.agent_b_model}
                    onChange={(e) => setConfig({ ...config, agent_b_model: e.target.value })}
                    className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                  >
                    {providers?.find(p => p.name === config.agent_b_provider)?.models.map((m) => (
                      <option key={m} value={m}>{m}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Persona</label>
                  <select
                    value={config.agent_b_persona}
                    onChange={(e) => setConfig({ ...config, agent_b_persona: e.target.value })}
                    className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                  >
                    {personas?.map((p) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="md:col-span-2 space-y-3 md:space-y-4 pt-4 border-t border-brand-border">
                 <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
                    <div>
                        <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Debate Style</label>
                        <select
                            value={config.style}
                            onChange={(e) => setConfig({ ...config, style: e.target.value })}
                            className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                        >
                            {styles?.map((s) => (
                            <option key={s.id} value={s.id}>{s.name}</option>
                            ))}
                        </select>
                    </div>
                    <div>
                        <label className="block text-xs md:text-sm font-medium text-[#859289] mb-1">Max Turns (per agent)</label>
                        <input
                            type="number"
                            min="1"
                            max="20"
                            value={config.max_turns}
                            onChange={(e) => setConfig({ ...config, max_turns: parseInt(e.target.value) })}
                            className="w-full bg-brand-bg border border-brand-border rounded-lg px-2 md:px-3 py-1.5 md:py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-xs md:text-sm"
                        />
                    </div>
                 </div>
              </div>
            </div>
          )}

          <div className="flex flex-col items-center space-y-4">
             <p className="text-sm text-[#859289]">
              Press Ctrl+Enter (⌘+Enter on Mac) to submit
            </p>
            <button
              type="submit"
              disabled={loading || !topic.trim() || !isReady}
              className="group inline-flex items-center px-12 py-5 border border-transparent rounded-xl text-xl font-bold text-[#2b3339] bg-brand-primary hover:bg-[#b8cc95] focus:outline-none focus:ring-4 focus:ring-brand-primary focus:ring-opacity-20 shadow-2xl transform transition-all duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
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
                  className="px-4 py-2 bg-brand-card border border-brand-border rounded-lg text-sm text-[#d3c6aa] hover:bg-brand-bg hover:border-brand-primary transition-colors"
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

import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import type { CreateCouncilRequest, MemberSpec } from '../types';

const EXAMPLE_TOPICS = [
  'How should we regulate AI safety?',
  'What is the future of space exploration?',
  'Should we adopt a 4-day work week?',
];

export function NewCouncil() {
  const navigate = useNavigate();
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [topic, setTopic] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Members state
  const [members, setMembers] = useState<MemberSpec[]>(() => {
    const saved = localStorage.getItem('council_members');
    return saved ? JSON.parse(saved) : [
      { Provider: '', Persona: '' },
      { Provider: '', Persona: '' },
    ];
  });

  const { data: providers } = useQuery({
    queryKey: ['providers'],
    queryFn: () => api.getProviders(),
  });

  const { data: personas } = useQuery({
    queryKey: ['personas'],
    queryFn: () => api.getPersonas(),
  });

  const { data: systemInfo } = useQuery({
    queryKey: ['systemInfo'],
    queryFn: () => api.getSystemInfo(),
  });

  const isReady = !!(providers?.length && personas?.length);

  // Auto-resize textarea
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  }, [topic]);

  // Set defaults when data loads
  useEffect(() => {
    if (providers?.length && personas?.length) {
      const availableProviders = providers.filter(p => p.available && p.name !== 'mock');
      const defaultProv = availableProviders.length > 0 ? availableProviders[0] : null;
      const defaultProviderName = defaultProv?.name || '';
      const defaultModelName = defaultProv?.default_model || '';
      
      // Initialize with default values if empty
      setMembers(prev => prev.map((m, i) => ({
        Provider: m.Provider || defaultProviderName,
        Model: m.Model || defaultModelName,
        Persona: m.Persona || (personas[i % personas.length]?.id || personas[0].id)
      })));
    }
  }, [providers, personas]);

  const handleAddMember = () => {
    if (!providers || !personas) return;
    const availableProviders = providers.filter(p => p.available && p.name !== 'mock');
    const defaultProv = availableProviders.length > 0 ? availableProviders[0] : null;
    const defaultProviderName = defaultProv?.name || '';
    const defaultModelName = defaultProv?.default_model || '';
    
    const newMembers = [...members, { 
      Provider: defaultProviderName, 
      Model: defaultModelName,
      Persona: personas[members.length % personas.length]?.id || personas[0].id 
    }];
    setMembers(newMembers);
    localStorage.setItem('council_members', JSON.stringify(newMembers));
  };

  const handleMemberChange = (index: number, field: keyof MemberSpec, value: string) => {
    const newMembers = [...members];
    if (field === 'Provider') {
      const provider = providers?.find(p => p.name === value);
      newMembers[index] = { 
        ...newMembers[index], 
        Provider: value, 
        Model: provider?.default_model || '' 
      };
    } else {
      newMembers[index] = { ...newMembers[index], [field]: value };
    }
    setMembers(newMembers);
    localStorage.setItem('council_members', JSON.stringify(newMembers));
  };

  const handleRemoveMember = (index: number) => {
    if (members.length <= 2) return;
    const newMembers = members.filter((_, i) => i !== index);
    setMembers(newMembers);
    localStorage.setItem('council_members', JSON.stringify(newMembers));
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
      // Validate members
      if (members.length < 2) {
        throw new Error('At least 2 members are required');
      }
      for (let i = 0; i < members.length; i++) {
        if (!members[i].Provider) {
          throw new Error(`Please select a provider for member ${i + 1}`);
        }
      }

      const request: CreateCouncilRequest = {
        Topic: topic.trim(),
        Members: members,
        auto_run: true,
      };

      const council = await api.createCouncil(request);
      navigate(`/councils/${council.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create council');
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
    <div className="min-h-[70vh] flex flex-col items-start justify-center py-12">
      <div className="max-w-6xl w-full px-4">
        <div className="mb-12 text-left">
          <div className="mb-6">
            <span className="text-5xl">🏛️</span>
          </div>
          <h1 className="text-4xl font-bold text-white mb-4 leading-tight">
            Convene a Council
          </h1>
          <p className="text-lg text-gray-400">
            Gather AI agents to deliberate, rank perspectives, and reach a consensus
          </p>
          {systemInfo?.cwd && (
            <div className="mt-6 flex items-center text-[#d3c6aa] text-xs font-mono bg-brand-bg bg-opacity-40 px-3 py-1.5 rounded border border-brand-border w-fit shadow-inner mx-auto lg:mx-0">
              <svg className="w-3.5 h-3.5 mr-1.5 text-brand-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
              <span className="opacity-60 mr-1 text-[#859289]">Running in:</span> {systemInfo.cwd}
            </div>
          )}
        </div>

        <form onSubmit={handleSubmit} className="space-y-8">
          {error && (
            <div className="animate-fadeIn bg-red-900 bg-opacity-50 border-2 border-red-500 text-red-200 px-6 py-4 rounded-xl shadow-lg">
              <div className="flex items-start">
                <svg className="w-6 h-6 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
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
              ref={textareaRef}
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
              required
              autoFocus
              placeholder="e.g., How should we regulate AI safety?"
              className="block w-full rounded-xl bg-brand-card border-2 border-brand-border text-[#d3c6aa] placeholder-[#5c6a72] focus:border-brand-primary focus:ring-4 focus:ring-brand-primary focus:ring-opacity-20 text-base p-6 transition-all duration-200 shadow-lg hover:shadow-2xl outline-none"
            />
          </div>

          <div className="bg-brand-card p-6 rounded-xl border border-brand-border animate-fadeIn space-y-4">
            <div className="flex justify-between items-center border-b border-brand-border pb-2">
              <h3 className="text-lg font-medium text-[#d3c6aa]">Council Members</h3>
              <span className="text-sm text-[#859289]">{members.length} members</span>
            </div>
            
            <div className="space-y-3">
              {members.map((member, index) => (
                <div key={index} className="flex gap-4 items-center bg-brand-bg p-3 rounded-lg border border-brand-border">
                  <div className="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-brand-card rounded-full text-[#859289] text-sm font-bold">
                    {index + 1}
                  </div>
                  
                  <div className="flex-1 grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <select
                        value={member.Provider}
                        onChange={(e) => handleMemberChange(index, 'Provider', e.target.value)}
                        className="w-full bg-brand-card border border-brand-border rounded-lg px-3 py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-sm"
                      >
                        {providers?.filter(p => p.name !== 'mock').map((p) => (
                          <option key={p.name} value={p.name} disabled={!p.available}>
                            {p.display_name} {!p.available && '(Unavailable)'}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <select
                        value={member.Model}
                        onChange={(e) => handleMemberChange(index, 'Model', e.target.value)}
                        className="w-full bg-brand-card border border-brand-border rounded-lg px-3 py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-sm"
                      >
                        {providers?.find(p => p.name === member.Provider)?.models.map((m) => (
                          <option key={m} value={m}>{m}</option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <select
                        value={member.Persona}
                        onChange={(e) => handleMemberChange(index, 'Persona', e.target.value)}
                        className="w-full bg-brand-card border border-brand-border rounded-lg px-3 py-2 text-[#d3c6aa] focus:ring-2 focus:ring-brand-primary text-sm"
                      >
                        {personas?.map((p) => (
                          <option key={p.id} value={p.id}>{p.name}</option>
                        ))}
                      </select>
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() => handleRemoveMember(index)}
                    disabled={members.length <= 2}
                    className="text-gray-500 hover:text-red-400 disabled:opacity-30 disabled:hover:text-gray-500"
                    title="Remove member"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>

            <div className="pt-2 flex justify-center">
              <button
                type="button"
                onClick={handleAddMember}
                className="flex items-center gap-2 px-4 py-2 bg-brand-card hover:bg-brand-bg rounded-lg text-sm text-[#d3c6aa] transition-colors border border-brand-border"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                Add Member
              </button>
            </div>
          </div>

          <div className="flex flex-col items-center space-y-4">
            <button
              type="submit"
              disabled={loading || !topic.trim() || !isReady}
              className="group inline-flex items-center px-12 py-5 border border-transparent rounded-xl text-xl font-bold text-[#2b3339] bg-brand-primary hover:bg-[#b8cc95] focus:outline-none focus:ring-4 focus:ring-brand-primary focus:ring-opacity-20 shadow-2xl transform transition-all duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
            >
              {loading ? (
                <>
                  <svg className="animate-spin -ml-1 mr-3 h-6 w-6 text-white" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  Convening...
                </>
              ) : (
                <>
                  <span className="group-hover:mr-1 transition-all duration-200">
                    Convene Council
                  </span>
                  <svg className="ml-2 w-6 h-6 group-hover:translate-x-1 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
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

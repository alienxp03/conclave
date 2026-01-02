import { useState, useEffect } from 'react';
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
  const [topic, setTopic] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Members state
  const [members, setMembers] = useState<MemberSpec[]>([
    { Provider: '', Persona: '' },
    { Provider: '', Persona: '' },
  ]);

  const { data: providers } = useQuery({
    queryKey: ['providers'],
    queryFn: () => api.getProviders(),
  });

  const { data: personas } = useQuery({
    queryKey: ['personas'],
    queryFn: () => api.getPersonas(),
  });

  const isReady = !!(providers?.length && personas?.length);

  // Set defaults when data loads
  useEffect(() => {
    if (providers?.length && personas?.length) {
      const availableProviders = providers.filter(p => p.available && p.name !== 'mock');
      const defaultProvider = availableProviders.length > 0 ? availableProviders[0].name : '';
      
      // Initialize with default values if empty
      setMembers(prev => prev.map((m, i) => ({
        Provider: m.Provider || defaultProvider,
        Persona: m.Persona || (personas[i % personas.length]?.id || personas[0].id)
      })));
    }
  }, [providers, personas]);

  const handleAddMember = () => {
    if (!providers || !personas) return;
    const availableProviders = providers.filter(p => p.available && p.name !== 'mock');
    const defaultProvider = availableProviders.length > 0 ? availableProviders[0].name : '';
    
    setMembers([...members, { 
      Provider: defaultProvider, 
      Persona: personas[members.length % personas.length]?.id || personas[0].id 
    }]);
  };

  const handleRemoveMember = (index: number) => {
    if (members.length <= 2) return;
    setMembers(members.filter((_, i) => i !== index));
  };

  const handleMemberChange = (index: number, field: keyof MemberSpec, value: string) => {
    const newMembers = [...members];
    newMembers[index] = { ...newMembers[index], [field]: value };
    setMembers(newMembers);
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
    <div className="min-h-[70vh] flex flex-col items-center justify-center py-12">
      <div className="max-w-4xl w-full mx-auto px-4">
        <div className="mb-12 text-center">
          <div className="mb-6">
            <span className="text-6xl">🏛️</span>
          </div>
          <h1 className="text-5xl font-bold text-white mb-4 leading-tight">
            Convene a Council
          </h1>
          <p className="text-xl text-gray-400">
            Gather AI agents to deliberate, rank perspectives, and reach a consensus
          </p>
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
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onKeyDown={handleKeyDown}
              rows={3}
              required
              autoFocus
              placeholder="e.g., How should we regulate AI safety?"
              className="block w-full rounded-xl bg-gray-800 border-2 border-gray-600 text-white placeholder-gray-500 focus:border-blue-500 focus:ring-4 focus:ring-blue-500 focus:ring-opacity-30 text-xl p-6 transition-all duration-200 shadow-lg hover:shadow-2xl outline-none"
            />
          </div>

          <div className="bg-gray-800 p-6 rounded-xl border border-gray-700 animate-fadeIn space-y-4">
            <div className="flex justify-between items-center border-b border-gray-700 pb-2">
              <h3 className="text-lg font-medium text-white">Council Members</h3>
              <span className="text-sm text-gray-400">{members.length} members</span>
            </div>
            
            <div className="space-y-3">
              {members.map((member, index) => (
                <div key={index} className="flex gap-4 items-center bg-gray-900 p-3 rounded-lg border border-gray-700">
                  <div className="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-gray-800 rounded-full text-gray-400 text-sm font-bold">
                    {index + 1}
                  </div>
                  
                  <div className="flex-1 grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <select
                        value={member.Provider}
                        onChange={(e) => handleMemberChange(index, 'Provider', e.target.value)}
                        className="w-full bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500 text-sm"
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
                        value={member.Persona}
                        onChange={(e) => handleMemberChange(index, 'Persona', e.target.value)}
                        className="w-full bg-gray-800 border border-gray-600 rounded-lg px-3 py-2 text-white focus:ring-2 focus:ring-blue-500 text-sm"
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
                className="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg text-sm text-white transition-colors border border-gray-600"
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
              className="group inline-flex items-center px-12 py-5 border border-transparent rounded-xl text-xl font-semibold text-white bg-gradient-to-r from-blue-600 to-blue-700 hover:from-blue-700 hover:to-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-500 focus:ring-opacity-50 shadow-2xl transform transition-all duration-200 hover:scale-105 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
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

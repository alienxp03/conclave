import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';

export function Projects() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const instructionsRef = useRef<HTMLTextAreaElement>(null);
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [instructions, setInstructions] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects', 'all'],
    queryFn: () => api.getProjects(100, 0),
  });

  useEffect(() => {
    if (!instructionsRef.current) return;
    instructionsRef.current.style.height = 'auto';
    instructionsRef.current.style.height = `${instructionsRef.current.scrollHeight}px`;
  }, [instructions, showForm]);

  const resetForm = () => {
    setName('');
    setDescription('');
    setInstructions('');
    setError(null);
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Project name is required.');
      return;
    }

    setSaving(true);
    setError(null);
    try {
      const project = await api.createProject({
        name: name.trim(),
        description: description.trim(),
        instructions,
      });
      await queryClient.invalidateQueries({ queryKey: ['projects'] });
      resetForm();
      navigate(`/projects/${project.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-10">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">Projects</h1>
          <p className="text-gray-400">Create shared context for related chats</p>
        </div>
        <button
          onClick={() => setShowForm(prev => !prev)}
          className="bg-brand-primary hover:bg-[#b8cc95] text-[#2b3339] px-6 py-3 rounded-xl text-base font-bold inline-flex items-center shadow-lg transform transition-all duration-200 hover:scale-105"
        >
          <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          {showForm ? 'Close' : 'New Project'}
        </button>
      </div>

      {showForm && (
        <form
          onSubmit={handleCreate}
          className="bg-brand-card border border-brand-border rounded-2xl p-6 space-y-4 animate-fadeIn"
        >
          {error && (
            <div className="bg-red-900/40 border border-red-500 text-red-200 px-4 py-3 rounded-xl text-sm">
              {error}
            </div>
          )}
          <div>
            <label className="text-xs uppercase tracking-widest text-[#859289] font-semibold">Project name</label>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="mt-2 w-full bg-brand-bg border border-brand-border rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-brand-primary"
              placeholder="New project"
              required
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-widest text-[#859289] font-semibold">Description</label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="mt-2 w-full bg-brand-bg border border-brand-border rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-brand-primary"
              placeholder="Optional short description"
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-widest text-[#859289] font-semibold">Project instructions</label>
            <textarea
              ref={instructionsRef}
              value={instructions}
              onChange={(e) => setInstructions(e.target.value)}
              rows={3}
              className="mt-2 w-full bg-brand-bg border border-brand-border rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-brand-primary resize-none"
              placeholder="Add project context or guidelines..."
            />
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={saving}
              className="bg-brand-primary text-[#2b3339] px-5 py-2.5 rounded-xl font-semibold shadow-lg hover:bg-[#b8cc95] transition-colors disabled:opacity-60"
            >
              {saving ? 'Creating...' : 'Create project'}
            </button>
            <button
              type="button"
              onClick={() => {
                resetForm();
                setShowForm(false);
              }}
              className="text-sm text-[#859289] hover:text-white transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {isLoading ? (
        <div className="text-gray-400">Loading projects...</div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          {(projects || []).map(project => (
            <Link key={project.id} to={`/projects/${project.id}`} className="group">
              <div className="bg-brand-card border border-brand-border rounded-2xl p-6 hover:border-brand-primary hover:shadow-xl transition-all duration-200">
                <h3 className="text-lg font-semibold text-white mb-2 group-hover:text-brand-primary transition-colors">
                  {project.name}
                </h3>
                <p className="text-sm text-[#859289]">
                  {project.description || 'No description'}
                </p>
                <div className="mt-4 text-xs text-[#5c6a72]">
                  Updated {new Date(project.updated_at).toLocaleString()}
                </div>
              </div>
            </Link>
          ))}
          {(projects || []).length === 0 && (
            <div className="text-sm text-[#859289]">No projects yet.</div>
          )}
        </div>
      )}
    </div>
  );
}

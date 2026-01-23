import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';

export function ProjectDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const instructionsRef = useRef<HTMLTextAreaElement>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [instructions, setInstructions] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['project', id],
    queryFn: () => api.getProject(id || ''),
    enabled: !!id,
  });

  useEffect(() => {
    if (!data?.project) return;
    setName(data.project.name);
    setDescription(data.project.description);
    setInstructions(data.project.instructions);
  }, [data?.project]);

  useEffect(() => {
    if (!instructionsRef.current) return;
    instructionsRef.current.style.height = 'auto';
    instructionsRef.current.style.height = `${instructionsRef.current.scrollHeight}px`;
  }, [instructions, isEditing]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;
    if (!name.trim()) {
      setError('Project name is required.');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const updated = await api.updateProject(id, {
        name: name.trim(),
        description: description.trim(),
        instructions,
      });
      await queryClient.invalidateQueries({ queryKey: ['projects'] });
      await queryClient.invalidateQueries({ queryKey: ['project', id] });
      setIsEditing(false);
      setName(updated.name);
      setDescription(updated.description);
      setInstructions(updated.instructions);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update project');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!id) return;
    if (!window.confirm('Delete this project and all its chats?')) return;
    try {
      await api.deleteProject(id);
      await queryClient.invalidateQueries({ queryKey: ['projects'] });
      navigate('/projects');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete project');
    }
  };

  if (isLoading || !data) {
    return <div className="text-gray-400">Loading project...</div>;
  }

  const { project, debates, councils } = data;
  const allChats = [
    ...(debates || []).map(d => ({ ...d, type: 'debate' as const })),
    ...(councils || []).map(c => ({ ...c, type: 'council' as const })),
  ].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

  return (
    <div className="space-y-10">
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold text-white mb-2">{project.name}</h1>
          <p className="text-gray-400 text-sm">Updated {new Date(project.updated_at).toLocaleString()}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link
            to={`/?project=${project.id}`}
            className="bg-brand-primary hover:bg-[#b8cc95] text-[#2b3339] px-4 py-2 rounded-xl text-sm font-bold inline-flex items-center shadow-lg transition-all duration-200"
          >
            Start new council
          </Link>
          <button
            onClick={() => setIsEditing(true)}
            className="px-4 py-2 rounded-xl text-sm border border-brand-border text-[#d3c6aa] hover:border-brand-primary transition-colors"
          >
            Edit project
          </button>
          <button
            onClick={handleDelete}
            className="px-4 py-2 rounded-xl text-sm border border-red-500/60 text-red-300 hover:border-red-400 transition-colors"
          >
            Delete
          </button>
        </div>
      </div>

      {isEditing && (
        <form onSubmit={handleSave} className="bg-brand-card border border-brand-border rounded-2xl p-6 space-y-4">
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
              required
            />
          </div>
          <div>
            <label className="text-xs uppercase tracking-widest text-[#859289] font-semibold">Description</label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="mt-2 w-full bg-brand-bg border border-brand-border rounded-xl px-4 py-3 text-white focus:outline-none focus:ring-2 focus:ring-brand-primary"
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
            />
          </div>
          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={saving}
              className="bg-brand-primary text-[#2b3339] px-5 py-2.5 rounded-xl font-semibold shadow-lg hover:bg-[#b8cc95] transition-colors disabled:opacity-60"
            >
              {saving ? 'Saving...' : 'Save changes'}
            </button>
            <button
              type="button"
              onClick={() => setIsEditing(false)}
              className="text-sm text-[#859289] hover:text-white transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      {!isEditing && (
        <div className="bg-brand-card border border-brand-border rounded-2xl p-6 text-sm text-[#859289]">
          <p>This project hides its instructions by default. Click “Edit project” to view or update them.</p>
        </div>
      )}

      <div>
        <h2 className="text-xl font-semibold text-white mb-4">Project chats</h2>
        {allChats.length === 0 ? (
          <div className="text-sm text-[#859289]">No chats yet.</div>
        ) : (
          <div className="grid gap-4">
            {allChats.map(chat => (
              <Link
                key={chat.id}
                to={chat.type === 'debate' ? `/debates/${chat.id}` : `/councils/${chat.id}`}
                className="bg-brand-bg border border-brand-border rounded-xl px-4 py-3 hover:border-brand-primary transition-colors"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <div className="text-white font-medium">{chat.title || chat.topic}</div>
                    <div className="text-xs text-[#5c6a72]">
                      {new Date(chat.created_at).toLocaleString()} • {chat.type}
                    </div>
                  </div>
                  <svg className="w-4 h-4 text-[#859289]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                  </svg>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

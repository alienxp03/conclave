import type { Debate, Provider, CreateDebateRequest, Turn, Persona, Style, Council, CouncilResponse, CouncilRanking, CreateCouncilRequest, CouncilSummary, SystemInfo, DebateStats, Project, DebateSummary } from '../types';

const API_BASE = '/api';

class ApiClient {
  async getSystemInfo(): Promise<SystemInfo> {
    const response = await fetch(`${API_BASE}/system/info`);
    if (!response.ok) throw new Error('Failed to fetch system info');
    return response.json();
  }

  async getProviders(): Promise<Provider[]> {
    const response = await fetch(`${API_BASE}/providers`);
    if (!response.ok) throw new Error('Failed to fetch providers');
    return response.json();
  }

  async getPersonas(): Promise<Persona[]> {
    const response = await fetch(`${API_BASE}/personas`);
    if (!response.ok) throw new Error('Failed to fetch personas');
    return response.json();
  }

  async getStyles(): Promise<Style[]> {
    const response = await fetch(`${API_BASE}/styles`);
    if (!response.ok) throw new Error('Failed to fetch styles');
    return response.json();
  }

  async getDebates(limit = 50, offset = 0): Promise<Debate[]> {
    const response = await fetch(`${API_BASE}/debates?limit=${limit}&offset=${offset}`);
    if (!response.ok) throw new Error('Failed to fetch debates');
    return response.json();
  }

  async getDebate(id: string): Promise<{ debate: Debate; turns: Turn[]; stats?: DebateStats }> {
    const response = await fetch(`${API_BASE}/debates/${id}`);
    if (!response.ok) throw new Error('Failed to fetch debate');
    return response.json();
  }

  async getProjects(limit = 20, offset = 0): Promise<Project[]> {
    const response = await fetch(`${API_BASE}/projects?limit=${limit}&offset=${offset}`);
    if (!response.ok) throw new Error('Failed to fetch projects');
    return response.json();
  }

  async getProject(id: string): Promise<{ project: Project; debates: DebateSummary[]; councils: CouncilSummary[] }> {
    const response = await fetch(`${API_BASE}/projects/${id}`);
    if (!response.ok) throw new Error('Failed to fetch project');
    return response.json();
  }

  async createProject(request: { name: string; description: string; instructions: string }): Promise<Project> {
    const response = await fetch(`${API_BASE}/projects`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to create project');
    }

    return response.json();
  }

  async updateProject(id: string, request: { name: string; description: string; instructions: string }): Promise<Project> {
    const response = await fetch(`${API_BASE}/projects/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to update project');
    }

    return response.json();
  }

  async deleteProject(id: string): Promise<void> {
    const response = await fetch(`${API_BASE}/projects/${id}`, {
      method: 'DELETE',
    });
    if (!response.ok) throw new Error('Failed to delete project');
  }

  async createDebate(request: CreateDebateRequest): Promise<Debate> {
    const response = await fetch(`${API_BASE}/debates`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to create debate');
    }

    return response.json();
  }

  async deleteDebate(id: string): Promise<void> {
    const response = await fetch(`${API_BASE}/debates/${id}`, {
      method: 'DELETE',
    });
    if (!response.ok) throw new Error('Failed to delete debate');
  }

  async addDebateFollowUp(id: string, content: string): Promise<void> {
    const response = await fetch(`${API_BASE}/debates/${id}/followup`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    });
    if (!response.ok) throw new Error('Failed to add follow-up');
  }

  // Create an EventSource for streaming debate updates
  createDebateStream(debateId: string): EventSource {
    return new EventSource(`${API_BASE}/debates/${debateId}/stream`);
  }

  // Council API
  async getCouncils(limit = 50, offset = 0): Promise<CouncilSummary[]> {
    const response = await fetch(`${API_BASE}/councils?limit=${limit}&offset=${offset}`);
    if (!response.ok) throw new Error('Failed to fetch councils');
    return response.json();
  }

  async getCouncil(id: string): Promise<{ council: Council; responses: CouncilResponse[]; rankings: CouncilRanking[] }> {
    const response = await fetch(`${API_BASE}/councils/${id}`);
    if (!response.ok) throw new Error('Failed to fetch council');
    return response.json();
  }

  async createCouncil(request: CreateCouncilRequest): Promise<Council> {
    const response = await fetch(`${API_BASE}/councils`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to create council');
    }

    return response.json();
  }

  async addCouncilFollowUp(id: string, content: string): Promise<void> {
    const response = await fetch(`${API_BASE}/councils/${id}/followup`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    });
    if (!response.ok) throw new Error('Failed to add follow-up');
  }

  createCouncilStream(councilId: string): EventSource {
    return new EventSource(`${API_BASE}/councils/${councilId}/stream`);
  }
}

export const api = new ApiClient();

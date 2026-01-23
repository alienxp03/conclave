import { render, screen, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { vi } from 'vitest';
import { Navigation } from '../Navigation';
import type { Debate, Project } from '../../types';
import { api } from '../../lib/api';

vi.mock('../../lib/api', () => ({
  api: {
    getDebates: vi.fn(),
    getCouncils: vi.fn(),
    getProjects: vi.fn(),
  },
}));

const mockedApi = vi.mocked(api, true);

const buildQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

describe('Navigation', () => {
  it('shows project labels under recent conversations', async () => {
    const debates: Debate[] = [
      {
        id: 'debate-1',
        title: 'Debate One',
        topic: 'Topic One',
        cwd: '/tmp',
        project_id: 'project-1',
        project_instructions: '',
        agent_a: {
          id: 'agent-a',
          name: 'Agent A',
          provider: 'provider',
          model: 'model-a',
          persona: 'persona-a',
        },
        agent_b: {
          id: 'agent-b',
          name: 'Agent B',
          provider: 'provider',
          model: 'model-b',
          persona: 'persona-b',
        },
        status: 'completed',
        style: 'default',
        total_turns: 0,
        turn_count: 0,
        read_only: false,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 'debate-2',
        title: 'Debate Two',
        topic: 'Topic Two',
        cwd: '/tmp',
        agent_a: {
          id: 'agent-a2',
          name: 'Agent A2',
          provider: 'provider',
          model: 'model-a2',
          persona: 'persona-a2',
        },
        agent_b: {
          id: 'agent-b2',
          name: 'Agent B2',
          provider: 'provider',
          model: 'model-b2',
          persona: 'persona-b2',
        },
        status: 'completed',
        style: 'default',
        total_turns: 0,
        turn_count: 0,
        read_only: false,
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      },
    ];

    const projects: Project[] = [
      {
        id: 'project-1',
        name: 'Project Alpha',
        description: '',
        instructions: '',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
    ];

    mockedApi.getDebates.mockResolvedValue(debates);
    mockedApi.getCouncils.mockResolvedValue([]);
    mockedApi.getProjects.mockResolvedValue(projects);

    const queryClient = buildQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/debates/debate-1']}>
          <Navigation isCollapsed={false} setIsCollapsed={vi.fn()} />
        </MemoryRouter>
      </QueryClientProvider>
    );

    const debateOneTitle = await screen.findByText('Debate One');
    const debateOneLink = debateOneTitle.closest('a');
    expect(debateOneLink).not.toBeNull();
    expect(within(debateOneLink as HTMLElement).getByText('Project Alpha')).toBeInTheDocument();

    const debateTwoTitle = await screen.findByText('Debate Two');
    const debateTwoLink = debateTwoTitle.closest('a');
    expect(debateTwoLink).not.toBeNull();
    expect(within(debateTwoLink as HTMLElement).queryByText('No project')).toBeNull();
  });
});

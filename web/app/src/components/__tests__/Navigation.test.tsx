import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
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
    updateDebateTitle: vi.fn(),
    updateCouncilTitle: vi.fn(),
    deleteDebate: vi.fn(),
    deleteCouncil: vi.fn(),
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
  beforeEach(() => {
    if (!window.prompt) {
      Object.defineProperty(window, 'prompt', {
        value: () => null,
        writable: true,
      });
    }
    if (!window.confirm) {
      Object.defineProperty(window, 'confirm', {
        value: () => false,
        writable: true,
      });
    }
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

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

  it('renames a debate from the sidebar menu', async () => {
    const debates: Debate[] = [
      {
        id: 'debate-1',
        title: 'Original Title',
        topic: 'Topic One',
        cwd: '/tmp',
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
    ];

    mockedApi.getDebates.mockResolvedValue(debates);
    mockedApi.getCouncils.mockResolvedValue([]);
    mockedApi.getProjects.mockResolvedValue([]);
    mockedApi.updateDebateTitle.mockResolvedValue();

    const promptSpy = vi.spyOn(window, 'prompt').mockReturnValue('New Title');
    const queryClient = buildQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/debates/debate-1']}>
          <Navigation isCollapsed={false} setIsCollapsed={vi.fn()} />
        </MemoryRouter>
      </QueryClientProvider>
    );

    const optionsButton = await screen.findByRole('button', { name: /conversation options/i });
    fireEvent.click(optionsButton);

    const renameButton = await screen.findByRole('button', { name: /rename title/i });
    fireEvent.click(renameButton);

    expect(promptSpy).toHaveBeenCalledWith('Rename conversation', 'Original Title');
    await waitFor(() => {
      expect(mockedApi.updateDebateTitle).toHaveBeenCalledWith('debate-1', 'New Title');
    });

    promptSpy.mockRestore();
  });

  it('deletes a council from the sidebar menu after confirmation', async () => {
    mockedApi.getDebates.mockResolvedValue([]);
    mockedApi.getCouncils.mockResolvedValue([
      {
        id: 'council-1',
        title: 'Council One',
        topic: 'Topic One',
        cwd: '/tmp',
        status: 'completed',
        member_count: 3,
        created_at: '2024-01-02T00:00:00Z',
      },
    ]);
    mockedApi.getProjects.mockResolvedValue([]);
    mockedApi.deleteCouncil.mockResolvedValue();

    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(true);
    const queryClient = buildQueryClient();

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/councils/council-1']}>
          <Navigation isCollapsed={false} setIsCollapsed={vi.fn()} />
        </MemoryRouter>
      </QueryClientProvider>
    );

    const optionsButton = await screen.findByRole('button', { name: /conversation options/i });
    fireEvent.click(optionsButton);

    const deleteButton = await screen.findByRole('button', { name: /delete conversation/i });
    fireEvent.click(deleteButton);

    expect(confirmSpy).toHaveBeenCalled();
    await waitFor(() => {
      expect(mockedApi.deleteCouncil).toHaveBeenCalledWith('council-1');
    });

    confirmSpy.mockRestore();
  });
});

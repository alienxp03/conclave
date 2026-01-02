export type DebateStatus = 'pending' | 'in_progress' | 'completed' | 'failed';

export interface Agent {
  id: string;
  name: string;
  provider: string;
  model: string;
  persona: string;
}

export interface Turn {
  id: string;
  debate_id: string;
  agent_id: string;
  number: number;
  content: string;
  created_at: string;
}

export interface Vote {
  agrees: boolean;
  reasoning?: string;
}

export interface Conclusion {
  agreed: boolean;
  early_consensus: boolean;
  summary: string;
  agent_a_summary?: string;
  agent_b_summary?: string;
  agent_a_vote?: Vote;
  agent_b_vote?: Vote;
}

export interface Debate {
  id: string;
  title: string;
  topic: string;
  cwd: string;
  agent_a: Agent;
  agent_b: Agent;
  status: DebateStatus;
  style: string;
  total_turns: number;
  turn_count: number;
  read_only: boolean;
  created_at: string;
  updated_at: string;
  completed_at?: string;
  conclusion?: Conclusion;
}

export interface DebateSummary {
  id: string;
  title: string;
  topic: string;
  cwd: string;
  status: DebateStatus;
  style: string;
  agent_a: string;
  agent_b: string;
  turn_count: number;
  read_only: boolean;
  created_at: string;
  type?: 'debate' | 'council';
}

export interface Provider {
  name: string;
  display_name: string;
  available: boolean;
}

export interface Persona {
  id: string;
  name: string;
  description: string;
  system_prompt: string;
}

export interface Style {
  id: string;
  name: string;
  description: string;
}

export interface CreateDebateRequest {
  topic: string;
  agent_a_provider: string;
  agent_a_model: string;
  agent_a_persona: string;
  agent_b_provider: string;
  agent_b_model: string;
  agent_b_persona: string;
  style: string;
  max_turns: number;
  auto_run?: boolean;
}

export interface CouncilSummary {
  id: string;
  title: string;
  topic: string;
  cwd: string;
  status: DebateStatus;
  member_count: number;
  created_at: string;
}

export interface StreamEvent {
  type: 'turn_start' | 'content' | 'turn_complete' | 'debate_complete' | 'error';
  data: any;
}

export interface MemberSpec {
  Provider: string;
  Model?: string;
  Persona?: string;
}

export interface CreateCouncilRequest {
  Topic: string;
  Members: MemberSpec[];
  Chairman?: MemberSpec;
  auto_run?: boolean;
}

export interface Council {
  id: string;
  title: string;
  topic: string;
  cwd: string;
  members: Agent[];
  chairman: Agent;
  status: DebateStatus;
  synthesis?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface SystemInfo {
  cwd: string;
}

export interface CouncilResponse {
  id: string;
  council_id: string;
  member_id: string;
  content: string;
  created_at: string;
}

export interface CouncilRanking {
  id: string;
  council_id: string;
  reviewer_id: string;
  rankings: string[];
  reasoning: string;
  created_at: string;
}

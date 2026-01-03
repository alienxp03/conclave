import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import type { Turn, Debate } from '../types';

interface TurnCardProps {
  turn: Turn;
  debate: Debate;
  isStreaming?: boolean;
}

export function TurnCard({ turn, debate, isStreaming = false }: TurnCardProps) {
  const isAgentA = turn.agent_id === debate.agent_a.id;
  const agent = isAgentA ? debate.agent_a : debate.agent_b;

  // Use explicit classes for Tailwind purge
  const avatarClasses = isAgentA
    ? 'w-12 h-12 rounded-full flex items-center justify-center shadow-lg bg-gradient-to-br from-brand-primary to-[#8fa173]'
    : 'w-12 h-12 rounded-full flex items-center justify-center shadow-lg bg-gradient-to-br from-brand-secondary to-[#c6ab70]';

  const nameClasses = isAgentA
    ? 'font-bold text-lg text-brand-primary'
    : 'font-bold text-lg text-brand-secondary';

  const borderClasses = isAgentA
    ? 'ml-15 bg-brand-card bg-opacity-40 rounded-xl p-6 border border-brand-primary border-opacity-20 group-hover:shadow-xl transition-shadow'
    : 'ml-15 bg-brand-card bg-opacity-40 rounded-xl p-6 border border-brand-secondary border-opacity-20 group-hover:shadow-xl transition-shadow';

  return (
    <div className={`animate-fadeIn group ${isStreaming ? 'opacity-80' : ''}`}>
      {/* Agent identifier */}
      <div className="flex items-center gap-3 mb-3">
        <div className={avatarClasses}>
          <span className="text-2xl">{isAgentA ? '💭' : '🧠'}</span>
        </div>
        <div className="flex-1">
          <div className="flex items-baseline gap-2">
            <span className={nameClasses}>
              {agent.name}
            </span>
            <span className="text-xs text-gray-500">
              Turn {turn.number}
              {!isStreaming && ` • ${new Date(turn.created_at).toLocaleString()}`}
              {isStreaming && ' • typing...'}
            </span>
          </div>
        </div>
      </div>

      {/* Argument content */}
      <div className={borderClasses}>
        <div className="text-[#d3c6aa] leading-relaxed text-base prose prose-invert prose-p:leading-relaxed prose-pre:bg-[#2b3339] prose-pre:border prose-pre:border-brand-border max-w-none">
          <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>
            {turn.content}
          </ReactMarkdown>
          {isStreaming && <span className="inline-block w-2 h-5 bg-gray-400 ml-1 animate-pulse" />}
        </div>
      </div>
    </div>
  );
}

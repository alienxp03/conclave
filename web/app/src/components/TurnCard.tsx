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
  const isFailed = turn.status === 'failed';

  // Use explicit classes for Tailwind purge
  const avatarClasses = isAgentA
    ? 'w-12 h-12 rounded-full flex items-center justify-center shadow-lg bg-gradient-to-br from-brand-primary to-[#8fa173]'
    : 'w-12 h-12 rounded-full flex items-center justify-center shadow-lg bg-gradient-to-br from-brand-secondary to-[#c6ab70]';

  const nameClasses = isAgentA
    ? 'font-bold text-lg text-brand-primary'
    : 'font-bold text-lg text-brand-secondary';

  const borderClasses = isFailed
    ? 'ml-15 bg-red-900 bg-opacity-20 rounded-xl p-6 border border-red-500 border-opacity-30 group-hover:shadow-xl transition-shadow'
    : isAgentA
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
            {isFailed && (
              <span className="text-xs text-red-500 font-semibold">• Failed</span>
            )}
          </div>
        </div>
      </div>

      {/* Argument content or error */}
      <div className={borderClasses}>
        {isFailed ? (
          <div className="text-red-300 leading-relaxed text-base">
            <div className="flex items-start gap-2 mb-2">
              <svg className="w-5 h-5 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div>
                <div className="font-semibold mb-1">Turn Failed</div>
                <div className="text-sm text-red-200 opacity-90">{turn.error || 'Unknown error'}</div>
                <div className="text-xs text-gray-400 mt-2">
                  The debate continues with remaining turns.
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="text-[#d3c6aa] leading-relaxed text-base prose prose-invert prose-p:leading-relaxed prose-pre:bg-[#2b3339] prose-pre:border prose-pre:border-brand-border max-w-none">
            <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>
              {turn.content}
            </ReactMarkdown>
            {isStreaming && <span className="inline-block w-2 h-5 bg-gray-400 ml-1 animate-pulse" />}
          </div>
        )}
      </div>
    </div>
  );
}

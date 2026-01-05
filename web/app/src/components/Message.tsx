import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import type { ReactNode } from 'react';

export type MessageRole = 'agent' | 'user' | 'system';

export interface MessageProps {
  role: MessageRole;
  name?: string;
  avatar?: string;
  agentColor?: 'primary' | 'secondary' | 'blue' | 'purple' | 'orange';
  isStreaming?: boolean;
  timestamp?: string;
  metadata?: string;
  children: ReactNode;
}

const markdownPlugins = [remarkGfm, remarkBreaks];

const colorClasses = {
  primary: {
    avatar: 'bg-gradient-to-br from-brand-primary to-[#8fa173]',
    name: 'text-brand-primary',
    border: 'border-brand-primary',
  },
  secondary: {
    avatar: 'bg-gradient-to-br from-brand-secondary to-[#c6ab70]',
    name: 'text-brand-secondary',
    border: 'border-brand-secondary',
  },
  blue: {
    avatar: 'bg-gradient-to-br from-brand-blue to-[#6ba39a]',
    name: 'text-brand-blue',
    border: 'border-brand-blue',
  },
  purple: {
    avatar: 'bg-gradient-to-br from-brand-purple to-[#c088a0]',
    name: 'text-brand-purple',
    border: 'border-brand-purple',
  },
  orange: {
    avatar: 'bg-gradient-to-br from-brand-orange to-[#cf8664]',
    name: 'text-brand-orange',
    border: 'border-brand-orange',
  },
};

function MessageRoot({ role, name, avatar, agentColor = 'primary', isStreaming, timestamp, metadata, children }: MessageProps) {
  if (role === 'system') {
    return (
      <div className="flex justify-center my-4 md:my-6">
        <div className="bg-brand-bg rounded-lg px-3 py-1.5 md:px-4 md:py-2 border border-brand-border">
          <div className="text-xs md:text-sm text-[#859289] font-medium">{children}</div>
        </div>
      </div>
    );
  }

  if (role === 'user') {
    return (
      <div className="flex gap-2 md:gap-3 mb-4 md:mb-6 animate-fadeIn">
        <div className="w-8 h-8 md:w-10 md:h-10 rounded-full flex items-center justify-center shadow-lg bg-gradient-to-br from-brand-accent to-[#d06c6e] flex-shrink-0">
          <span className="text-lg md:text-xl">👤</span>
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-baseline gap-2 mb-1 md:mb-2">
            <span className="font-semibold text-sm md:text-base text-white">
              {name || 'You'}
            </span>
            {timestamp && (
              <span className="text-[10px] md:text-xs text-[#859289]">
                {new Date(timestamp).toLocaleString()}
              </span>
            )}
          </div>
          <div className="bg-brand-card bg-opacity-60 rounded-xl p-3 md:p-4 border border-brand-border">
            <div className="text-[#d3c6aa] leading-relaxed text-sm md:text-base prose prose-invert prose-p:leading-relaxed prose-pre:bg-[#2b3339] prose-pre:border prose-pre:border-brand-border max-w-none">
              {typeof children === 'string' ? (
                <ReactMarkdown remarkPlugins={markdownPlugins}>{children}</ReactMarkdown>
              ) : (
                children
              )}
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Agent message
  const colors = colorClasses[agentColor];

  return (
    <div className={`flex gap-2 md:gap-3 mb-4 md:mb-6 animate-fadeIn ${isStreaming ? 'opacity-90' : ''}`}>
      <div className={`w-8 h-8 md:w-10 md:h-10 rounded-full flex items-center justify-center shadow-lg ${colors.avatar} flex-shrink-0`}>
        <span className="text-lg md:text-xl">{avatar || '🤖'}</span>
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline gap-2 mb-1 md:mb-2">
          <span className={`font-semibold text-sm md:text-base ${colors.name}`}>
            {name || 'Agent'}
          </span>
          {metadata && (
            <span className="text-[10px] md:text-xs text-[#859289]">{metadata}</span>
          )}
          {timestamp && (
            <span className="text-[10px] md:text-xs text-[#859289]">
              {isStreaming ? 'typing...' : new Date(timestamp).toLocaleString()}
            </span>
          )}
        </div>
        <div className={`bg-brand-card bg-opacity-40 rounded-xl p-3 md:p-4 border ${colors.border} border-opacity-20 hover:shadow-xl transition-shadow`}>
          <div className="text-[#d3c6aa] leading-relaxed text-sm md:text-base prose prose-invert prose-p:leading-relaxed prose-pre:bg-[#2b3339] prose-pre:border prose-pre:border-brand-border max-w-none">
            {typeof children === 'string' ? (
              <ReactMarkdown remarkPlugins={markdownPlugins}>{children}</ReactMarkdown>
            ) : (
              children
            )}
            {isStreaming && <span className="inline-block w-2 h-4 md:h-5 bg-[#859289] ml-1 animate-pulse" />}
          </div>
        </div>
      </div>
    </div>
  );
}

// Composable subcomponents following assistant-ui pattern
function MessageContent({ children }: { children: ReactNode }) {
  return (
    <div className="text-[#d3c6aa] leading-relaxed text-base prose prose-invert prose-p:leading-relaxed prose-pre:bg-[#2b3339] prose-pre:border prose-pre:border-brand-border max-w-none">
      {typeof children === 'string' ? (
        <ReactMarkdown remarkPlugins={markdownPlugins}>{children}</ReactMarkdown>
      ) : (
        children
      )}
    </div>
  );
}

function MessageAvatar({ emoji, color = 'primary' }: { emoji: string; color?: keyof typeof colorClasses }) {
  const colors = colorClasses[color];
  return (
    <div className={`w-10 h-10 rounded-full flex items-center justify-center shadow-lg ${colors.avatar}`}>
      <span className="text-xl">{emoji}</span>
    </div>
  );
}

export const Message = {
  Root: MessageRoot,
  Content: MessageContent,
  Avatar: MessageAvatar,
};

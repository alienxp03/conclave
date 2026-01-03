import type { ReactNode } from 'react';

interface RoundContainerProps {
  roundNumber?: number;
  title?: string;
  stage?: string;
  children: ReactNode;
}

export function RoundContainer({
  roundNumber,
  title,
  stage,
  children
}: RoundContainerProps) {
  return (
    <div className="mb-8">
      {/* Round header */}
      {(roundNumber !== undefined || title || stage) && (
        <div className="sticky top-0 z-10 bg-brand-bg border-b border-brand-border mb-6 pb-3">
          <div className="flex items-center gap-3">
            {roundNumber !== undefined && (
              <div className="flex items-center justify-center w-8 h-8 rounded-full bg-brand-primary text-brand-bg font-bold text-sm">
                {roundNumber}
              </div>
            )}
            <div className="flex-1">
              {title && (
                <h3 className="text-lg font-semibold text-white">{title}</h3>
              )}
              {stage && (
                <p className="text-sm text-[#859289] mt-0.5">{stage}</p>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Messages in this round */}
      <div className="space-y-4">
        {children}
      </div>
    </div>
  );
}

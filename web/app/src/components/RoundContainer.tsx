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
    <div className="mb-6 md:mb-8">
      {/* Round header */}
      {(roundNumber !== undefined || title || stage) && (
        <div className="sticky top-0 z-10 bg-brand-bg border-b border-brand-border mb-4 md:mb-6 pb-2 md:pb-3">
          <div className="flex items-center gap-2 md:gap-3">
            {roundNumber !== undefined && (
              <div className="flex items-center justify-center w-6 h-6 md:w-8 md:h-8 rounded-full bg-brand-primary text-brand-bg font-bold text-xs md:text-sm flex-shrink-0">
                {roundNumber}
              </div>
            )}
            <div className="flex-1">
              {title && (
                <h3 className="text-base md:text-lg font-semibold text-white">{title}</h3>
              )}
              {stage && (
                <p className="text-xs md:text-sm text-[#859289] mt-0.5">{stage}</p>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Messages in this round */}
      <div className="space-y-3 md:space-y-4">
        {children}
      </div>
    </div>
  );
}

import { useEffect, useState } from 'react';
import type { Turn, Debate } from '../types';

interface StreamingTurn {
  agentId: string;
  number: number;
  round: number;
  content: string;
}

export function useDebateStream(debateId: string | undefined, initialTurns: Turn[] = []) {
  const [turns, setTurns] = useState<Turn[]>(initialTurns);
  const [streamingTurn, setStreamingTurn] = useState<StreamingTurn | null>(null);
  const [debate, setDebate] = useState<Debate | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Update turns if initialTurns changes (e.g. after a query refetch)
  useEffect(() => {
    if (initialTurns.length > turns.length) {
      setTurns(initialTurns);
    }
  }, [initialTurns]);

  useEffect(() => {
    if (!debateId) return;

    const eventSource = new EventSource(`/api/debates/${debateId}/stream`);

    eventSource.addEventListener('turn_start', (e) => {
      const data = JSON.parse(e.data);
      setStreamingTurn({
        agentId: data.agent_id,
        number: data.number,
        round: data.round || 1,
        content: '',
      });
    });

    eventSource.addEventListener('content', (e) => {
      setStreamingTurn((prev) => {
        if (!prev) return null;
        return {
          ...prev,
          content: prev.content + e.data,
        };
      });
    });

    eventSource.addEventListener('turn_complete', (e) => {
      const turn: Turn = JSON.parse(e.data);
      setTurns((prev) => [...prev, turn]);
      setStreamingTurn(null);
    });

    eventSource.addEventListener('debate_complete', (e) => {
      const updatedDebate: Debate = JSON.parse(e.data);
      setDebate(updatedDebate);
      setStreamingTurn(null);
    });

    eventSource.addEventListener('error', (e: any) => {
      if (e.data) {
        const errorData = JSON.parse(e.data);
        setError(errorData.message);
      }
    });

    eventSource.onerror = () => {
      eventSource.close();
    };

    return () => {
      eventSource.close();
    };
  }, [debateId]);

  return { turns, streamingTurn, debate, error };
}

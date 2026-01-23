import { useState, useEffect, useCallback } from 'react';

export interface ProviderHealthStatus {
  available: boolean;
  response_time: number;
  error?: string;
  checked_at: string;
}

export interface ProvidersHealth {
  [providerName: string]: ProviderHealthStatus;
}

export function useProviderHealth() {
  const [health, setHealth] = useState<ProvidersHealth>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const checkHealth = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/providers/health');
      if (!response.ok) {
        throw new Error(`Health check failed: ${response.statusText}`);
      }

      const data = await response.json();
      setHealth(data.providers || {});
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to check provider health');
      console.error('Provider health check failed:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  // Run health check on mount
  useEffect(() => {
    checkHealth();
  }, [checkHealth]);

  return {
    health,
    loading,
    error,
    refresh: checkHealth,
  };
}

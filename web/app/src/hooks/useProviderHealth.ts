import { useState, useEffect, useCallback } from 'react';

export interface ProviderHealthStatus {
  available: boolean;
  response_time: number;
  error?: string;
  checked_at: string;
}

export interface ProviderInfo {
  name: string;
  display_name?: string;
  default_model?: string;
  available: boolean;
}

export interface ProvidersHealth {
  [providerName: string]: ProviderHealthStatus;
}

export function useProviderHealth() {
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [health, setHealth] = useState<ProvidersHealth>({});
  const [loadingProviders, setLoadingProviders] = useState(true);
  const [checkingProviders, setCheckingProviders] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);

  const checkProvider = useCallback(async (providerName: string, force = false) => {
    setCheckingProviders((prev) => {
      const next = new Set(prev);
      next.add(providerName);
      return next;
    });

    try {
      const query = force ? '?force=1' : '';
      const response = await fetch(`/api/providers/health/${encodeURIComponent(providerName)}${query}`);
      if (!response.ok) {
        throw new Error(`Health check failed: ${response.statusText}`);
      }

      const data = await response.json();
      setHealth((prev) => ({
        ...prev,
        [providerName]: {
          available: Boolean(data.available),
          response_time: Number(data.response_time ?? 0),
          error: data.error || undefined,
          checked_at: data.checked_at,
        },
      }));
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to check provider health';
      setHealth((prev) => ({
        ...prev,
        [providerName]: {
          available: false,
          response_time: 0,
          error: message,
          checked_at: new Date().toISOString(),
        },
      }));
      console.error('Provider health check failed:', err);
    } finally {
      setCheckingProviders((prev) => {
        const next = new Set(prev);
        next.delete(providerName);
        return next;
      });
    }
  }, []);

  const refresh = useCallback(() => {
    providers.forEach((provider) => {
      void checkProvider(provider.name, true);
    });
  }, [providers, checkProvider]);

  useEffect(() => {
    let cancelled = false;

    const fetchProviders = async () => {
      setLoadingProviders(true);
      setError(null);

      try {
        const response = await fetch('/api/providers');
        if (!response.ok) {
          throw new Error(`Provider list failed: ${response.statusText}`);
        }

        const data = await response.json();
        if (cancelled) {
          return;
        }

        const list = Array.isArray(data) ? data : [];
        setProviders(list);
        setLoadingProviders(false);

        list.forEach((provider) => {
          void checkProvider(provider.name);
        });
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load providers');
          setLoadingProviders(false);
        }
      }
    };

    void fetchProviders();

    return () => {
      cancelled = true;
    };
  }, [checkProvider]);

  return {
    providers,
    health,
    loadingProviders,
    checkingProviders,
    error,
    refresh,
  };
}

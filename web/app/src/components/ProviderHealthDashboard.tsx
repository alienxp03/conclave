import { useProviderHealth } from '../hooks/useProviderHealth';

export function ProviderHealthDashboard() {
  const {
    providers,
    health,
    loadingProviders,
    checkingProviders,
    error,
    refresh,
  } = useProviderHealth();

  const getStatusColor = (status: any) => {
    if (!status.available) return 'text-red-500';
    if (status.response_time > 5) return 'text-yellow-500';
    if (status.response_time > 3) return 'text-yellow-400';
    return 'text-green-500';
  };

  const getStatusColorForName = (name: string) => {
    const status = health[name];
    if (!status) return 'text-brand-text-secondary';
    return getStatusColor(status);
  };

  const getStatusText = (status: any) => {
    if (!status.available) return 'Unavailable';
    if (status.response_time > 5) return 'Slow';
    return 'Available';
  };

  const getStatusTextForName = (name: string) => {
    const status = health[name];
    if (!status) return checkingProviders.has(name) || loadingProviders ? 'Checking...' : 'Unknown';
    return getStatusText(status);
  };

  const formatResponseTime = (seconds: number) => {
    return seconds < 1 ? `${(seconds * 1000).toFixed(0)}ms` : `${seconds.toFixed(1)}s`;
  };

  const formatCheckedAt = (checkedAt?: string) => {
    if (!checkedAt) return '—';
    const date = new Date(checkedAt);
    if (Number.isNaN(date.getTime())) return '—';
    return date.toLocaleString();
  };

  const formatModelName = (model?: string, providerName?: string) => {
    if (!model) return '';
    let normalized = model;
    if (providerName && normalized.startsWith(`${providerName}-`)) {
      normalized = normalized.slice(providerName.length + 1);
    }
    normalized = normalized.replace(/-(\d)-(\d)$/, ' $1.$2');
    return normalized.replace(/[_-]+/g, ' ').trim();
  };

  return (
    <div className="bg-brand-card border border-brand-border rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-brand-text">Provider Status</h3>
        <button
          onClick={refresh}
          disabled={loadingProviders || checkingProviders.size > 0}
          className="text-xs text-brand-primary hover:text-brand-primary-dark transition-colors disabled:opacity-50"
        >
          {loadingProviders || checkingProviders.size > 0 ? 'Checking...' : 'Refresh'}
        </button>
      </div>

      {error && (
        <div className="text-xs text-red-500 mb-2">{error}</div>
      )}

      <div className="space-y-2">
        {providers.map((provider) => {
          const name = provider.name;
          const status = health[name];
          return (
            <div key={name} className="flex items-center justify-between text-sm">
              <div className="flex items-center space-x-2">
                <span className={`${getStatusColorForName(name)} text-lg`}>●</span>
                <span className="text-brand-text capitalize">
                  {provider.display_name || name}
                  {provider.default_model ? ` (${formatModelName(provider.default_model, name)})` : ''}
                </span>
              </div>
              <div className="flex items-center space-x-3 text-brand-text-secondary">
                <span className={getStatusColorForName(name)}>
                  {getStatusTextForName(name)}
                </span>
                {status?.available && (
                  <span className="text-xs">
                    {formatResponseTime(status.response_time)}
                  </span>
                )}
                <span className="text-xs">
                  Last checked {formatCheckedAt(status?.checked_at)}
                </span>
                {status?.error && (
                  <span className="text-xs text-red-500 max-w-xs truncate" title={status.error}>
                    {status.error}
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {providers.length === 0 && !loadingProviders && (
        <div className="text-sm text-brand-text-secondary">No providers configured</div>
      )}
      {providers.length === 0 && loadingProviders && (
        <div className="text-sm text-brand-text-secondary">Loading providers...</div>
      )}
    </div>
  );
}

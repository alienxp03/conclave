import { useProviderHealth } from '../hooks/useProviderHealth';

export function ProviderHealthDashboard() {
  const { health, loading, error, refresh } = useProviderHealth();

  if (loading && Object.keys(health).length === 0) {
    return (
      <div className="bg-brand-card border border-brand-border rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-brand-text">Provider Status</h3>
        </div>
        <div className="text-sm text-brand-text-secondary">Checking provider health...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-brand-card border border-brand-border rounded-lg p-4 mb-4">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-brand-text">Provider Status</h3>
          <button
            onClick={refresh}
            className="text-xs text-brand-primary hover:text-brand-primary-dark transition-colors"
          >
            Refresh
          </button>
        </div>
        <div className="text-sm text-red-500">Failed to check provider health</div>
      </div>
    );
  }

  const getStatusColor = (status: any) => {
    if (!status.available) return 'text-red-500';
    if (status.response_time > 5) return 'text-yellow-500';
    if (status.response_time > 3) return 'text-yellow-400';
    return 'text-green-500';
  };

  const getStatusText = (status: any) => {
    if (!status.available) return 'Unavailable';
    if (status.response_time > 5) return 'Slow';
    return 'Available';
  };

  const formatResponseTime = (seconds: number) => {
    return seconds < 1 ? `${(seconds * 1000).toFixed(0)}ms` : `${seconds.toFixed(1)}s`;
  };

  return (
    <div className="bg-brand-card border border-brand-border rounded-lg p-4 mb-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-medium text-brand-text">Provider Status</h3>
        <button
          onClick={refresh}
          disabled={loading}
          className="text-xs text-brand-primary hover:text-brand-primary-dark transition-colors disabled:opacity-50"
        >
          {loading ? 'Checking...' : 'Refresh'}
        </button>
      </div>

      <div className="space-y-2">
        {Object.entries(health).map(([name, status]) => (
          <div key={name} className="flex items-center justify-between text-sm">
            <div className="flex items-center space-x-2">
              <span className={`${getStatusColor(status)} text-lg`}>●</span>
              <span className="text-brand-text capitalize">{name}</span>
            </div>
            <div className="flex items-center space-x-3 text-brand-text-secondary">
              <span className={getStatusColor(status)}>
                {getStatusText(status)}
              </span>
              {status.available && (
                <span className="text-xs">
                  {formatResponseTime(status.response_time)}
                </span>
              )}
              {status.error && (
                <span className="text-xs text-red-500 max-w-xs truncate" title={status.error}>
                  {status.error}
                </span>
              )}
            </div>
          </div>
        ))}
      </div>

      {Object.keys(health).length === 0 && !loading && (
        <div className="text-sm text-brand-text-secondary">No providers configured</div>
      )}
    </div>
  );
}

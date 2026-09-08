const PUBLIC_MONITOR_TYPE_LABELS: Record<string, string> = {
  http: 'HTTP(s)',
  http_keyword: 'HTTP Keyword',
  http_json: 'HTTP JSON',
  tcp: 'TCP Port',
  ping: 'Ping',
  dns: 'DNS'
};

export function publicMonitorTypeLabel(type?: string): string {
  const normalized = String(type || '').trim().toLowerCase();
  return PUBLIC_MONITOR_TYPE_LABELS[normalized] || (normalized ? normalized.toUpperCase() : '—');
}

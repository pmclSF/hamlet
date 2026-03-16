export function logEvent(actor: string, action: string, resource: string) {
  return { eventId: 'evt_' + Date.now(), actor, action, resource, timestamp: new Date().toISOString() };
}

export function queryAuditLog(orgId: string, filters: any) {
  return { orgId, events: [], total: 0 };
}

export function exportAuditLog(orgId: string, format: string) {
  return { orgId, format, url: 'https://export.example.com/audit.csv' };
}

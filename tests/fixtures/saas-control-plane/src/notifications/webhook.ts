export function registerWebhook(orgId: string, url: string) {
  return { webhookId: 'wh_' + Date.now(), orgId, url, active: true };
}

export function triggerWebhook(webhookId: string, payload: any) {
  return { webhookId, delivered: true, statusCode: 200 };
}

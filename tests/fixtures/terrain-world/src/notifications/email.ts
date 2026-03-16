export function sendEmail(to: string, subject: string, body: string) {
  return { messageId: 'msg_' + Date.now(), to, subject, status: 'sent' };
}

export function sendBulkEmail(recipients: string[], subject: string) {
  return recipients.map(to => ({ to, status: 'queued' }));
}

export function sendAlert(recipientId: string, message: string) {
  return { alertId: 'alt_' + Date.now(), recipientId, message, status: 'sent' };
}
export function sendAppointmentReminder(appointmentId: string) {
  return { appointmentId, reminded: true };
}
export function sendLabResultNotification(patientId: string, labOrderId: string) {
  return { patientId, labOrderId, notified: true };
}

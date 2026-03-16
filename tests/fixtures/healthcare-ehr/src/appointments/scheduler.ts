import { getPatient } from '../patients/registry';
export function scheduleAppointment(patientId: string, doctorId: string, date: string) {
  const patient = getPatient(patientId);
  return { appointmentId: 'apt_' + Date.now(), patientId, doctorId, date, status: 'scheduled' };
}
export function cancelAppointment(appointmentId: string) {
  return { appointmentId, status: 'cancelled' };
}
export function rescheduleAppointment(appointmentId: string, newDate: string) {
  return { appointmentId, date: newDate, status: 'rescheduled' };
}

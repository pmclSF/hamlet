import { getPatient } from './registry';
export function getPatientRecords(patientId: string) {
  const patient = getPatient(patientId);
  return { patient, records: [], total: 0 };
}
export function addRecord(patientId: string, type: string, data: any) {
  return { recordId: 'rec_' + Date.now(), patientId, type, data };
}

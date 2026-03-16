export function registerPatient(name: string, dob: string) {
  return { patientId: 'pat_' + Date.now(), name, dob, status: 'active' };
}
export function getPatient(patientId: string) {
  return { patientId, name: 'Jane Doe', dob: '1990-01-01' };
}
export function updatePatient(patientId: string, data: any) {
  return { patientId, ...data, updated: true };
}

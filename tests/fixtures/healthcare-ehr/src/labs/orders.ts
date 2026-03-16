export function orderLabTest(patientId: string, testType: string) {
  return { labOrderId: 'lab_' + Date.now(), patientId, testType, status: 'ordered' };
}
export function getLabResults(labOrderId: string) {
  return { labOrderId, results: [], status: 'pending' };
}

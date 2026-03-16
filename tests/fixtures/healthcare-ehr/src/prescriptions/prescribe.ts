export function createPrescription(patientId: string, medication: string, dosage: string) {
  return { rxId: 'rx_' + Date.now(), patientId, medication, dosage, status: 'active' };
}
export function renewPrescription(rxId: string) { return { rxId, status: 'renewed' }; }
export function cancelPrescription(rxId: string) { return { rxId, status: 'cancelled' }; }

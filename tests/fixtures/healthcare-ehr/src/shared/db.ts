export function connectDB() { return { connected: true }; }
export function seedPatient() { return { id: 'pat_test', name: 'Test Patient' }; }
export function seedDoctor() { return { id: 'doc_test', name: 'Dr. Test' }; }
export function cleanupDB() { return { cleaned: true }; }
export function getRecord(id: string) { return { id }; }
export function seedAppointment() { return { id: 'apt_test' }; }
export function seedPrescription() { return { id: 'rx_test' }; }
export function seedLabOrder() { return { id: 'lab_test' }; }

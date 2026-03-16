import { describe, it, expect } from 'vitest';
import { scheduleAppointment, cancelAppointment, rescheduleAppointment } from '../../../src/appointments/scheduler';
describe('scheduleAppointment', () => {
  it('should schedule', () => { expect(scheduleAppointment('pat_1', 'doc_1', '2026-04-01').status).toBe('scheduled'); });
});
describe('cancelAppointment', () => {
  it('should cancel', () => { expect(cancelAppointment('apt_1').status).toBe('cancelled'); });
});

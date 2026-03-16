import { describe, it, expect } from 'vitest';
import { assessTriage, classifySymptoms } from '../../../src/ai-triage/model';
import { loadTriageDataset } from '../../../src/ai-triage/dataset';
describe('triage accuracy', () => {
  it('should assess triage', () => { expect(assessTriage('chest pain').triageLevel).toBeDefined(); });
  it('should classify symptoms', () => { expect(classifySymptoms('cough fever').category).toBe('respiratory'); });
  it('should load dataset', () => { expect(loadTriageDataset().length).toBeGreaterThan(0); });
});

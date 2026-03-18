import { describe, it, expect } from 'vitest';
import { safetyOverlay } from '../../../src/safety/guardrails';
import { policyBlock } from '../../../src/contexts/system';
describe('content safety', () => {
  it('should enforce policy', () => { expect(policyBlock).toContain('Never'); });
  it('should have safety overlay', () => { expect(safetyOverlay).toContain('harmful'); });
});

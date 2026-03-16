import { describe, it, expect } from 'vitest';
import { analyzeTransaction } from '../../../src/fraud/detector';
import { evaluateRules } from '../../../src/fraud/rules';
import { connectDB, getTransaction, cleanupDB } from '../../../src/shared-db';

describe('fraud integration', () => {
  it('should analyze and evaluate rules', () => {
    connectDB();
    const tx = getTransaction('tx_1');
    const analysis = analyzeTransaction(tx.id, tx.amount);
    const rules = evaluateRules(tx);
    expect(analysis.riskScore).toBeDefined();
    expect(rules.length).toBeGreaterThan(0);
    cleanupDB();
  });
});

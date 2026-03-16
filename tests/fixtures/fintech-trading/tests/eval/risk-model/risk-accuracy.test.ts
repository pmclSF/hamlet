import { describe, it, expect } from 'vitest';
import { predictRisk, classifyAsset } from '../../../src/risk/model';
import { loadTradingDataset } from '../../../src/risk/dataset';
describe('risk accuracy', () => {
  it('should predict', () => { expect(predictRisk('test').riskLevel).toBeDefined(); });
  it('should classify', () => { expect(classifyAsset('AAPL').assetClass).toBe('equity'); });
  it('should load data', () => { expect(loadTradingDataset().length).toBeGreaterThan(0); });
});

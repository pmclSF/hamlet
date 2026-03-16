import { describe, it, expect } from 'vitest';
import { predict, classify } from '../../../src/ai-recommendations/model';
import { loadEvalDataset } from '../../../src/ai-recommendations/dataset';
describe('model accuracy', () => { it('predict', () => { expect(predict('test').confidence).toBeGreaterThan(0); }); it('classify', () => { expect(classify('test').category).toBeDefined(); }); it('data', () => { expect(loadEvalDataset().length).toBeGreaterThan(0); }); });

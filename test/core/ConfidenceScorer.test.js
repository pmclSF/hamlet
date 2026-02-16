import { ConfidenceScorer } from '../../src/core/ConfidenceScorer.js';
import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  MockCall,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
} from '../../src/core/ir.js';

describe('ConfidenceScorer', () => {
  let scorer;

  beforeEach(() => {
    scorer = new ConfidenceScorer();
  });

  describe('score', () => {
    it('should return 100% when all nodes are converted', () => {
      const assertion1 = new Assertion({ kind: 'equal', confidence: 'converted' });
      const assertion2 = new Assertion({ kind: 'truthy', confidence: 'converted' });
      const tc = new TestCase({ name: 'test', body: [assertion1, assertion2], confidence: 'converted' });
      const suite = new TestSuite({ name: 'suite', tests: [tc], confidence: 'converted' });
      const file = new TestFile({ body: [suite] });

      const report = scorer.score(file);
      expect(report.confidence).toBe(100);
      expect(report.level).toBe('high');
      expect(report.converted).toBe(4);
      expect(report.unconvertible).toBe(0);
      expect(report.warnings).toBe(0);
    });

    it('should return 0% when all weighted nodes are unconvertible', () => {
      const assertion = new Assertion({ kind: 'equal', confidence: 'unconvertible' });
      const tc = new TestCase({ name: 'test', body: [assertion], confidence: 'unconvertible' });
      const file = new TestFile({ body: [tc] });

      const report = scorer.score(file);
      expect(report.confidence).toBe(0);
      expect(report.level).toBe('low');
    });

    it('should compute correct percentage for mixed results', () => {
      // 7 converted (weight 3+3+2+2+2+2+2 = 16) vs 3 unconvertible (weight 2+2+2 = 6)
      // Actually let's be precise:
      // Suite(3, converted) + TestCase(3, converted) + 4 assertions(2 each):
      //   2 converted (weight 4) + 2 unconvertible (weight 4)
      // Total weight = 3+3+4+4 = 14, converted weight = 3+3+4 = 10
      // Score = 10/14 * 100 = 71%
      const a1 = new Assertion({ kind: 'equal', confidence: 'converted' });
      const a2 = new Assertion({ kind: 'truthy', confidence: 'converted' });
      const a3 = new Assertion({ kind: 'deepEqual', confidence: 'unconvertible', sourceLocation: { line: 10 } });
      const a4 = new Assertion({ kind: 'throws', confidence: 'unconvertible', sourceLocation: { line: 15 } });
      const tc = new TestCase({ name: 'test', body: [a1, a2, a3, a4], confidence: 'converted' });
      const suite = new TestSuite({ name: 'suite', tests: [tc], confidence: 'converted' });
      const file = new TestFile({ body: [suite] });

      const report = scorer.score(file);
      // suite(3) + tc(3) + a1(2) + a2(2) = 10 converted
      // a3(2) + a4(2) = 4 unconvertible
      // total = 14, score = 10/14 = 71%
      expect(report.confidence).toBe(71);
      expect(report.level).toBe('medium');
      expect(report.converted).toBe(4);
      expect(report.unconvertible).toBe(2);
    });

    it('should count warnings as converted weight but track separately', () => {
      const a1 = new Assertion({ kind: 'equal', confidence: 'converted' });
      const a2 = new Assertion({ kind: 'truthy', confidence: 'warning', sourceLocation: { line: 5 } });
      const tc = new TestCase({ name: 'test', body: [a1, a2], confidence: 'converted' });
      const file = new TestFile({ body: [tc] });

      const report = scorer.score(file);
      // tc(3) + a1(2) + a2(2) = 7 converted weight, total = 7
      expect(report.confidence).toBe(100);
      expect(report.warnings).toBe(1);
      expect(report.converted).toBe(2); // tc + a1
    });

    it('should produce detail lines for unconvertible nodes', () => {
      const a = new Assertion({
        kind: 'snapshot',
        confidence: 'unconvertible',
        sourceLocation: { line: 23 },
        originalSource: 'expect(tree).toMatchSnapshot()',
      });
      const tc = new TestCase({ name: 'test', body: [a], confidence: 'converted' });
      const file = new TestFile({ body: [tc] });

      const report = scorer.score(file);
      expect(report.details).toHaveLength(1);
      expect(report.details[0].type).toBe('unconvertible');
      expect(report.details[0].line).toBe(23);
      expect(report.details[0].source).toBe('expect(tree).toMatchSnapshot()');
    });

    it('should produce detail lines for warning nodes', () => {
      const mock = new MockCall({
        kind: 'mockModule',
        confidence: 'warning',
        sourceLocation: { line: 5 },
        originalSource: 'jest.mock("./render")',
      });
      const tc = new TestCase({ name: 'test', body: [mock], confidence: 'converted' });
      const file = new TestFile({ body: [tc] });

      const report = scorer.score(file);
      expect(report.details).toHaveLength(1);
      expect(report.details[0].type).toBe('warning');
      expect(report.details[0].line).toBe(5);
    });

    it('should return 100% for empty test file', () => {
      const file = new TestFile();
      const report = scorer.score(file);
      expect(report.confidence).toBe(100);
      expect(report.total).toBe(0);
    });

    it('should not count RawCode or Comment nodes in weight', () => {
      const raw = new RawCode({ code: 'const x = 42;', confidence: 'converted' });
      const comment = new Comment({ text: '// TODO', confidence: 'converted' });
      const tc = new TestCase({ name: 'test', body: [raw, comment], confidence: 'converted' });
      const file = new TestFile({ body: [tc] });

      const report = scorer.score(file);
      // Only tc(3) has weight, raw and comment have 0 weight
      expect(report.confidence).toBe(100);
      expect(report.converted).toBe(1); // only tc
    });

    it('should weight structural nodes higher than assertions', () => {
      // One unconvertible suite (weight 3) vs one converted assertion (weight 2)
      const a = new Assertion({ kind: 'equal', confidence: 'converted' });
      const suite = new TestSuite({
        name: 'suite',
        tests: [new TestCase({ name: 'test', body: [a], confidence: 'converted' })],
        confidence: 'unconvertible',
      });
      const file = new TestFile({ body: [suite] });

      const report = scorer.score(file);
      // suite(3, unconvertible) + tc(3, converted) + a(2, converted) = total 8, converted 5
      // 5/8 = 62.5% → 63%
      expect(report.confidence).toBe(63);
    });

    it('should include import statements with weight 1', () => {
      const imp = new ImportStatement({
        source: 'vitest',
        confidence: 'converted',
      });
      const tc = new TestCase({ name: 'test', confidence: 'converted' });
      const file = new TestFile({ imports: [imp], body: [tc] });

      const report = scorer.score(file);
      // imp(1) + tc(3) = 4 total, all converted
      expect(report.confidence).toBe(100);
      expect(report.converted).toBe(2);
    });

    it('should report total count as sum of converted + unconvertible + warning', () => {
      const a1 = new Assertion({ confidence: 'converted' });
      const a2 = new Assertion({ confidence: 'unconvertible' });
      const a3 = new Assertion({ confidence: 'warning' });
      const tc = new TestCase({ body: [a1, a2, a3], confidence: 'converted' });
      const file = new TestFile({ body: [tc] });

      const report = scorer.score(file);
      expect(report.total).toBe(4); // tc + a1 + a2 + a3
      expect(report.converted).toBe(2); // tc + a1
      expect(report.unconvertible).toBe(1); // a2
      expect(report.warnings).toBe(1); // a3
    });
  });

  describe('getLevel', () => {
    it('should return high for 90-100', () => {
      expect(scorer.getLevel(100)).toBe('high');
      expect(scorer.getLevel(90)).toBe('high');
      expect(scorer.getLevel(95)).toBe('high');
    });

    it('should return medium for 70-89', () => {
      expect(scorer.getLevel(89)).toBe('medium');
      expect(scorer.getLevel(70)).toBe('medium');
      expect(scorer.getLevel(75)).toBe('medium');
    });

    it('should return low for below 70', () => {
      expect(scorer.getLevel(69)).toBe('low');
      expect(scorer.getLevel(0)).toBe('low');
      expect(scorer.getLevel(50)).toBe('low');
    });
  });
});

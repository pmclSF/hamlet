import { ConversionPipeline } from '../../src/core/ConversionPipeline.js';
import { FrameworkRegistry } from '../../src/core/FrameworkRegistry.js';
import {
  TestFile,
  TestSuite,
  TestCase,
  Assertion,
  ImportStatement,
  RawCode,
} from '../../src/core/ir.js';

function makeFramework(overrides = {}) {
  return {
    name: 'testfw',
    language: 'javascript',
    paradigm: 'bdd',
    detect: () => 80,
    parse: (source) => new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [
            new TestCase({
              name: 'test',
              body: [new Assertion({ kind: 'equal', confidence: 'converted' })],
              confidence: 'converted',
            }),
          ],
          confidence: 'converted',
        }),
      ],
    }),
    emit: (_ir, _source) => 'converted code',
    imports: {},
    ...overrides,
  };
}

describe('ConversionPipeline', () => {
  let registry;
  let pipeline;

  beforeEach(() => {
    registry = new FrameworkRegistry();
    pipeline = new ConversionPipeline(registry);
  });

  describe('convert', () => {
    it('should call detect → parse → transform → emit → score in order', async () => {
      const callOrder = [];

      const source = makeFramework({
        name: 'source',
        detect: (code) => { callOrder.push('detect'); return 80; },
        parse: (code) => {
          callOrder.push('parse');
          return new TestFile({
            body: [new TestCase({ name: 'test', confidence: 'converted' })],
          });
        },
      });

      const target = makeFramework({
        name: 'target',
        emit: (ir) => { callOrder.push('emit'); return 'output'; },
      });

      registry.register(source);
      registry.register(target);

      await pipeline.convert('input code', 'source', 'target');

      expect(callOrder).toEqual(['detect', 'parse', 'emit']);
    });

    it('should return { code, report } with correct shape', async () => {
      const source = makeFramework({ name: 'source' });
      const target = makeFramework({
        name: 'target',
        emit: () => 'converted output',
      });

      registry.register(source);
      registry.register(target);

      const result = await pipeline.convert('input', 'source', 'target');

      expect(result).toHaveProperty('code');
      expect(result).toHaveProperty('report');
      expect(result.code).toBe('converted output');
      expect(result.report).toHaveProperty('confidence');
      expect(result.report).toHaveProperty('converted');
      expect(result.report).toHaveProperty('unconvertible');
      expect(result.report).toHaveProperty('warnings');
      expect(result.report).toHaveProperty('details');
      expect(result.report).toHaveProperty('level');
    });

    it('should pass source code to parse', async () => {
      let parsedCode = null;

      const source = makeFramework({
        name: 'source',
        parse: (code) => {
          parsedCode = code;
          return new TestFile();
        },
      });
      const target = makeFramework({ name: 'target' });

      registry.register(source);
      registry.register(target);

      await pipeline.convert('my test code', 'source', 'target');
      expect(parsedCode).toBe('my test code');
    });

    it('should pass IR to emit', async () => {
      let emittedIr = null;
      const ir = new TestFile({
        body: [new TestCase({ name: 'test', confidence: 'converted' })],
      });

      const source = makeFramework({
        name: 'source',
        parse: () => ir,
      });
      const target = makeFramework({
        name: 'target',
        emit: (receivedIr) => { emittedIr = receivedIr; return 'output'; },
      });

      registry.register(source);
      registry.register(target);

      await pipeline.convert('input', 'source', 'target');
      expect(emittedIr).toBe(ir);
    });

    it('should compute confidence from IR node flags', async () => {
      const source = makeFramework({
        name: 'source',
        parse: () => new TestFile({
          body: [
            new TestCase({
              name: 'test1',
              confidence: 'converted',
              body: [
                new Assertion({ kind: 'equal', confidence: 'converted' }),
                new Assertion({ kind: 'snapshot', confidence: 'unconvertible' }),
              ],
            }),
          ],
        }),
      });
      const target = makeFramework({ name: 'target' });

      registry.register(source);
      registry.register(target);

      const result = await pipeline.convert('input', 'source', 'target');
      expect(result.report.confidence).toBeLessThan(100);
      expect(result.report.unconvertible).toBeGreaterThan(0);
    });

    it('should throw for unknown source framework', async () => {
      const target = makeFramework({ name: 'target' });
      registry.register(target);

      await expect(
        pipeline.convert('input', 'nonexistent', 'target')
      ).rejects.toThrow(/Unknown source framework.*nonexistent/);
    });

    it('should throw for unknown target framework', async () => {
      const source = makeFramework({ name: 'source' });
      registry.register(source);

      await expect(
        pipeline.convert('input', 'source', 'nonexistent')
      ).rejects.toThrow(/Unknown target framework.*nonexistent/);
    });

    it('should throw when detection confidence is 0 for non-empty input', async () => {
      const source = makeFramework({ name: 'source', detect: () => 0 });
      const target = makeFramework({ name: 'target' });

      registry.register(source);
      registry.register(target);

      await expect(
        pipeline.convert('some code', 'source', 'target')
      ).rejects.toThrow(/does not appear to be source/);
    });

    it('should handle empty source gracefully', async () => {
      const source = makeFramework({
        name: 'source',
        detect: () => 0,
        parse: () => new TestFile(),
      });
      const target = makeFramework({
        name: 'target',
        emit: () => '',
      });

      registry.register(source);
      registry.register(target);

      // Empty string should not throw (detection confidence 0 is OK for empty input)
      const result = await pipeline.convert('', 'source', 'target');
      expect(result.code).toBe('');
      expect(result.report.confidence).toBe(100);
    });

    it('should handle source with no recognizable test patterns', async () => {
      const source = makeFramework({
        name: 'source',
        detect: () => 10,
        parse: () => new TestFile({
          body: [new RawCode({ code: 'const x = 42;' })],
        }),
      });
      const target = makeFramework({
        name: 'target',
        emit: () => 'const x = 42;',
      });

      registry.register(source);
      registry.register(target);

      const result = await pipeline.convert('const x = 42;', 'source', 'target');
      expect(result.report.confidence).toBe(100); // RawCode is converted
      expect(result.report.total).toBeGreaterThanOrEqual(0);
    });

    it('should accept language option for disambiguation', async () => {
      const jsFw = makeFramework({ name: 'selenium', language: 'javascript' });
      const pyFw = makeFramework({ name: 'selenium', language: 'python' });
      const target = makeFramework({ name: 'target', language: 'javascript' });

      registry.register(jsFw);
      registry.register(pyFw);
      registry.register(target);

      let usedFramework = null;
      jsFw.parse = () => { usedFramework = 'js'; return new TestFile(); };
      pyFw.parse = () => { usedFramework = 'py'; return new TestFile(); };

      await pipeline.convert('code', 'selenium', 'target', { language: 'javascript' });
      expect(usedFramework).toBe('js');
    });
  });

  describe('transform', () => {
    it('should pass through IR unchanged for same paradigm', () => {
      const ir = new TestFile({ body: [new TestCase({ name: 'test' })] });
      const sourceFw = { paradigm: 'bdd' };
      const targetFw = { paradigm: 'bdd' };

      const result = pipeline.transform(ir, sourceFw, targetFw);
      expect(result).toBe(ir);
    });

    it('should return IR for different paradigms (stub — no transform yet)', () => {
      const ir = new TestFile({ body: [new TestCase({ name: 'test' })] });
      const sourceFw = { paradigm: 'bdd' };
      const targetFw = { paradigm: 'xunit' };

      const result = pipeline.transform(ir, sourceFw, targetFw);
      expect(result).toBe(ir);
    });
  });
});

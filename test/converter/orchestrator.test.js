import { ConversionOrchestrator } from '../../src/converter/orchestrator.js';

describe('ConversionOrchestrator', () => {
  let orchestrator;

  beforeEach(() => {
    orchestrator = new ConversionOrchestrator({
      converter: async (content) => content,
      configConverter: async () => 'config',
      validateTests: false,
      compareVisuals: false,
      generateTypes: false
    });
  });

  describe('constructor', () => {
    it('should initialize with provided options', () => {
      expect(orchestrator.options.validateTests).toBe(false);
      expect(orchestrator.options.compareVisuals).toBe(false);
      expect(orchestrator.options.generateTypes).toBe(false);
    });

    it('should initialize stats', () => {
      expect(orchestrator.stats.totalFiles).toBe(0);
      expect(orchestrator.stats.convertedFiles).toBe(0);
      expect(orchestrator.stats.skippedFiles).toBe(0);
      expect(orchestrator.stats.errors).toEqual([]);
    });

    it('should use default options when not provided', () => {
      const defaultOrch = new ConversionOrchestrator({});
      expect(defaultOrch.options.validateTests).toBe(true);
      expect(defaultOrch.options.compareVisuals).toBe(false);
      expect(defaultOrch.options.generateTypes).toBe(true);
    });

    it('should initialize components', () => {
      expect(orchestrator.validator).toBeDefined();
      expect(orchestrator.pluginConverter).toBeDefined();
      expect(orchestrator.visualComparison).toBeDefined();
      expect(orchestrator.typeScriptConverter).toBeDefined();
      expect(orchestrator.testMapper).toBeDefined();
    });
  });

  describe('generateReport', () => {
    it('should include statistics', () => {
      orchestrator.stats.totalFiles = 10;
      orchestrator.stats.convertedFiles = 8;
      orchestrator.stats.skippedFiles = 2;
      orchestrator.stats.startTime = new Date(Date.now() - 5000);
      orchestrator.stats.endTime = new Date();

      const report = orchestrator.generateReport();
      expect(report.statistics.totalFiles).toBe(10);
      expect(report.statistics.convertedFiles).toBe(8);
      expect(report.statistics.skippedFiles).toBe(2);
      expect(report.statistics.successRate).toBeDefined();
      expect(report.timestamp).toBeDefined();
    });
  });

  describe('getExecutionTime', () => {
    it('should return N/A when times not set', () => {
      expect(orchestrator.getExecutionTime()).toBe('N/A');
    });

    it('should return formatted time when both times set', () => {
      orchestrator.stats.startTime = new Date(Date.now() - 2500);
      orchestrator.stats.endTime = new Date();
      const time = orchestrator.getExecutionTime();
      expect(time).toContain('seconds');
    });
  });
});

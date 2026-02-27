import {
  getTargetExtension,
  buildOutputFilename,
  countTodos,
} from '../../src/cli/outputHelpers.js';

describe('outputHelpers', () => {
  describe('getTargetExtension', () => {
    it('should return .cy.js for cypress', () => {
      expect(getTargetExtension('cypress', '.js')).toBe('.cy.js');
    });

    it('should return .spec.js for playwright', () => {
      expect(getTargetExtension('playwright', '.js')).toBe('.spec.js');
    });

    it('should return .test.js for other frameworks', () => {
      expect(getTargetExtension('jest', '.js')).toBe('.test.js');
    });

    it('should preserve .py extension', () => {
      expect(getTargetExtension('pytest', '.py')).toBe('.py');
    });

    it('should preserve .java extension', () => {
      expect(getTargetExtension('junit5', '.java')).toBe('.java');
    });

    it('should default to .js when no original extension', () => {
      expect(getTargetExtension('vitest', '')).toBe('.test.js');
    });
  });

  describe('buildOutputFilename', () => {
    it('should replace .test suffix for vitest', () => {
      expect(buildOutputFilename('auth.test.js', 'vitest')).toBe(
        'auth.test.js'
      );
    });

    it('should replace .cy suffix for playwright', () => {
      expect(buildOutputFilename('auth.cy.js', 'playwright')).toBe(
        'auth.spec.js'
      );
    });

    it('should replace .spec suffix for cypress', () => {
      expect(buildOutputFilename('auth.spec.js', 'cypress')).toBe(
        'auth.cy.js'
      );
    });

    it('should preserve .py files', () => {
      expect(buildOutputFilename('test_auth.py', 'pytest')).toBe(
        'test_auth.py'
      );
    });
  });

  describe('countTodos', () => {
    it('should count HAMLET-TODO markers', () => {
      expect(
        countTodos('// HAMLET-TODO: fix this\n// HAMLET-TODO: and this')
      ).toBe(2);
    });

    it('should return 0 for no markers', () => {
      expect(countTodos('no todos here')).toBe(0);
    });
  });
});

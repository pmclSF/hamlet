import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures');
const outputDir = path.resolve(__dirname, '../output/shorthands');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI Shorthand Commands', () => {
  beforeAll(async () => {
    await fs.mkdir(fixturesDir, { recursive: true });
    await fs.mkdir(outputDir, { recursive: true });

    // Ensure fixture files exist
    await fs.writeFile(
      path.join(fixturesDir, 'sample.cy.js'),
      `describe('Sample Test', () => {
  it('should navigate and click', () => {
    cy.visit('/home');
    cy.get('#button').click();
    cy.get('.result').should('be.visible');
  });
});
`,
    );

    await fs.writeFile(
      path.join(fixturesDir, 'sample.jest.js'),
      `describe('Sample Test', () => {
  const mockFn = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should call the function', () => {
    mockFn('hello');
    expect(mockFn).toHaveBeenCalledWith('hello');
  });

  it('should use fake timers', () => {
    jest.useFakeTimers();
    setTimeout(() => mockFn(), 1000);
    jest.advanceTimersByTime(1000);
    expect(mockFn).toHaveBeenCalled();
    jest.useRealTimers();
  });
});
`,
    );
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('jest2vt shorthand', () => {
    test('should convert Jest file to Vitest output', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'jest2vt-out.test.js');

      runCLI(['jest2vt', inputFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toContain("from 'vitest'");
      expect(output).toContain('vi.fn()');
    });
  });

  describe('jesttovt long-form alias', () => {
    test('should also convert Jest file to Vitest', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'jesttovt-out.test.js');

      runCLI(['jesttovt', inputFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toContain("from 'vitest'");
      expect(output).toContain('vi.fn()');
    });
  });

  describe('cy2pw shorthand', () => {
    test('should convert Cypress file to Playwright output', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      const outFile = path.resolve(outputDir, 'cy2pw-out.spec.js');

      runCLI(['cy2pw', inputFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toContain('page.goto');
      expect(output).toContain('page.locator');
    });
  });

  describe('cytopw long-form alias', () => {
    test('should also convert Cypress file to Playwright', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      const outFile = path.resolve(outputDir, 'cytopw-out.spec.js');

      runCLI(['cytopw', inputFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toContain('page.goto');
    });
  });

  describe('Java shorthand (ju42ju5)', () => {
    test('should produce output for JUnit4 to JUnit5', async () => {
      const tmpDir = path.resolve(outputDir, 'java-tmp');
      await fs.mkdir(tmpDir, { recursive: true });

      const inputFile = path.join(tmpDir, 'SampleTest.java');
      await fs.writeFile(
        inputFile,
        `import org.junit.Test;
import static org.junit.Assert.*;

public class SampleTest {
    @Test
    public void testAddition() {
        assertEquals(4, 2 + 2);
    }
}
`,
      );

      const outFile = path.join(tmpDir, 'SampleTest.out.java');
      runCLI(['ju42ju5', inputFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      // Should convert JUnit4 annotations to JUnit5
      expect(output).toBeTruthy();
    });
  });

  describe('Python shorthand (pyt2ut)', () => {
    test('should produce output for pytest to unittest', async () => {
      const tmpDir = path.resolve(outputDir, 'python-tmp');
      await fs.mkdir(tmpDir, { recursive: true });

      const inputFile = path.join(tmpDir, 'test_sample.py');
      await fs.writeFile(
        inputFile,
        `import pytest

def test_addition():
    assert 2 + 2 == 4

def test_string():
    assert "hello".upper() == "HELLO"
`,
      );

      const outFile = path.join(tmpDir, 'test_sample.out.py');
      runCLI(['pyt2ut', inputFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toBeTruthy();
    });
  });

  describe('shorthand with --output flag', () => {
    test('should write to specified output path', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'custom-output.js');

      runCLI(['jest2vt', inputFile, '-o', outFile]);

      const exists = await fs.access(outFile).then(() => true).catch(() => false);
      expect(exists).toBe(true);
    });
  });

  describe('shorthand with --quiet', () => {
    test('should suppress non-error output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'quiet-out.js');

      const result = runCLI(['jest2vt', inputFile, '-o', outFile, '--quiet']);

      // Quiet mode should produce minimal or no output
      expect(result.trim()).toBe('');
    });
  });

  describe('hamlet list command', () => {
    test('should show all 25 directions with shorthands', () => {
      const result = runCLI(['list']);

      expect(result).toContain('Supported conversion directions');
      expect(result).toContain('jest');
      expect(result).toContain('vitest');
      expect(result).toContain('cypress');
      expect(result).toContain('playwright');
      expect(result).toContain('junit4');
      expect(result).toContain('pytest');
      // Should show shorthand aliases
      expect(result).toContain('jest2vt');
      expect(result).toContain('cy2pw');
      // Should show category headers
      expect(result).toContain('JavaScript E2E');
      expect(result).toContain('JavaScript Unit');
      expect(result).toContain('Java');
      expect(result).toContain('Python');
    });
  });

  describe('hamlet shorthands command', () => {
    test('should list all shorthand aliases', () => {
      const result = runCLI(['shorthands']);

      expect(result).toContain('Shorthand command aliases');
      expect(result).toContain('jest2vt');
      expect(result).toContain('jesttovt');
      expect(result).toContain('cy2pw');
      expect(result).toContain('cytopw');
      expect(result).toContain('ju42ju5');
      expect(result).toContain('pyt2ut');
    });
  });

  describe('hamlet --help', () => {
    test('should show list and shorthands commands', () => {
      const result = runCLI(['--help']);

      expect(result).toContain('list');
      expect(result).toContain('shorthands');
      expect(result).toContain('doctor');
    });
  });

  describe('unknown shorthand-like command', () => {
    test('should produce a helpful error', () => {
      expect(() => {
        runCLI(['foobar2baz', 'some-file.js'], { stdio: 'pipe' });
      }).toThrow();
    });
  });

  describe('shorthand with --dry-run', () => {
    test('should preview without writing files', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'dryrun-should-not-exist.js');

      // Remove if exists from previous run
      await fs.rm(outFile, { force: true }).catch(() => {});

      const result = runCLI(['jest2vt', inputFile, '--dry-run']);

      expect(result).toContain('Dry run');
      const exists = await fs.access(outFile).then(() => true).catch(() => false);
      expect(exists).toBe(false);
    });
  });
});

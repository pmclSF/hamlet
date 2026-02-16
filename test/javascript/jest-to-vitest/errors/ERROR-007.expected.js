import { describe, it, expect, afterEach } from 'vitest';

describe('TempFileManager', () => {
  let tempFiles = [];

  afterEach(() => {
    const errors = [];
    for (const file of tempFiles) {
      try {
        deleteFile(file);
      } catch (error) {
        errors.push(`Failed to delete ${file}: ${error.message}`);
      }
    }
    tempFiles = [];
    if (errors.length > 0) {
      console.warn('Cleanup warnings:', errors.join(', '));
    }
  });

  it('should create a temporary file', () => {
    const file = createTempFile('test-data.txt');
    tempFiles.push(file);
    expect(file).toBeDefined();
    expect(fileExists(file)).toBe(true);
  });

  it('should write and read temp file content', () => {
    const file = createTempFile('content-test.txt');
    tempFiles.push(file);
    writeToFile(file, 'hello world');
    const content = readFile(file);
    expect(content).toBe('hello world');
  });
});

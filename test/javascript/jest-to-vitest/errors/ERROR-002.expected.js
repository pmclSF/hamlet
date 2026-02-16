import { describe, it, expect } from 'vitest';

describe('FileReader', () => {
  it('should throw with message for missing file', () => {
    expect(() => readFileSync('/nonexistent/path.txt')).toThrow('File not found');
  });

  it('should throw with path in error message', () => {
    expect(() => readFileSync('/restricted/secret.key')).toThrow(/permission denied/i);
  });

  it('should throw a descriptive message for unsupported encoding', () => {
    expect(() => readFileSync('/data/file.bin', { encoding: 'utf-99' }))
      .toThrow('Unsupported encoding: utf-99');
  });

  it('should include the file size limit in the error', () => {
    expect(() => readFileSync('/data/huge.dat', { maxSize: 1024 }))
      .toThrow(/exceeds maximum size of 1024 bytes/);
  });
});

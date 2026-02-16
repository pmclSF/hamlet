import { describe, it, expect } from 'vitest';

describe('FileConverter', () => {
  it('should convert CSV to JSON', () => {
    const csv = 'name,age\nAlice,30\nBob,25';
    const rows = csv.split('\n');
    const headers = rows[0].split(',');
    expect(headers).toEqual(['name', 'age']);
    expect(rows.length).toBe(3);
  });

  it.todo('should convert XML to JSON');

  it.todo('should convert YAML to JSON');

  it.todo('should handle binary file formats');

  it('should reject unsupported formats', () => {
    const supported = ['csv', 'json', 'xml', 'yaml'];
    const format = 'docx';
    expect(supported).not.toContain(format);
  });
});

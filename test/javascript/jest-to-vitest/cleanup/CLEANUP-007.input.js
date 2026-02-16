const os = require('os');
const path = require('path');
const fs = require('fs');

describe('TempFileCleanup', () => {
  let tempDir;
  let tempFiles;

  beforeEach(() => {
    tempDir = os.tmpdir();
    tempFiles = [];
  });

  afterEach(() => {
    tempFiles.forEach((filePath) => {
      try {
        fs.unlinkSync(filePath);
      } catch (err) {
        // File may already be deleted
      }
    });
    tempFiles = [];
  });

  it('should create and track temp files', () => {
    const filePath = path.join(tempDir, `test-${Date.now()}.tmp`);
    fs.writeFileSync(filePath, 'temp data');
    tempFiles.push(filePath);

    expect(fs.existsSync(filePath)).toBe(true);
    expect(fs.readFileSync(filePath, 'utf8')).toBe('temp data');
  });

  it('should handle multiple temp files', () => {
    const file1 = path.join(tempDir, `test-a-${Date.now()}.tmp`);
    const file2 = path.join(tempDir, `test-b-${Date.now()}.tmp`);
    fs.writeFileSync(file1, 'data1');
    fs.writeFileSync(file2, 'data2');
    tempFiles.push(file1, file2);

    expect(tempFiles).toHaveLength(2);
  });
});

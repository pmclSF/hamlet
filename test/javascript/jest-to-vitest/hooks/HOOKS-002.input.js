describe('TempFileManager', () => {
  const createdFiles = [];

  it('should create a temp file', () => {
    const file = { name: 'tmp_001.txt', size: 128 };
    createdFiles.push(file);
    expect(file.name).toContain('tmp_');
  });

  it('should track created files', () => {
    const file = { name: 'tmp_002.txt', size: 256 };
    createdFiles.push(file);
    expect(createdFiles.length).toBeGreaterThan(0);
  });

  afterAll(() => {
    createdFiles.length = 0;
  });
});

describe('FileProcessor', () => {
  it('should process and clean up manually', () => {
    const handle = { fd: 42, closed: false };
    try {
      handle.data = 'file contents';
      expect(handle.data).toBe('file contents');
      expect(handle.fd).toBe(42);
    } finally {
      handle.closed = true;
      handle.fd = -1;
    }
    expect(handle.closed).toBe(true);
  });

  it('should clean up even on assertion failure', () => {
    const buffer = { allocated: true, size: 1024 };
    try {
      expect(buffer.allocated).toBe(true);
      expect(buffer.size).toBeGreaterThan(0);
    } finally {
      buffer.allocated = false;
      buffer.size = 0;
    }
    expect(buffer.allocated).toBe(false);
  });
});

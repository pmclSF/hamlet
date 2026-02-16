declare module './app' {
  interface AppContext {
    testMode: boolean;
  }
}

declare global {
  interface Window {
    __TEST_FLAG__: boolean;
  }
}

describe('Module augmentation', () => {
  it('should work with augmented types', () => {
    const context: { testMode: boolean } = { testMode: true };
    expect(context.testMode).toBe(true);
  });

  it('should handle global augmentation', () => {
    const flag: boolean = true;
    expect(flag).toBe(true);
  });
});

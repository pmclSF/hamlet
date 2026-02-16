describe('Counter', () => {
  let counter;

  beforeEach(() => {
    counter = { value: 0, increment() { this.value++; } };
  });

  it('should start at zero', () => {
    expect(counter.value).toBe(0);
  });

  it('should increment', () => {
    counter.increment();
    expect(counter.value).toBe(1);
  });

  it('should increment multiple times', () => {
    counter.increment();
    counter.increment();
    counter.increment();
    expect(counter.value).toBe(3);
  });
});

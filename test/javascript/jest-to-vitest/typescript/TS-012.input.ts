function getName(): string {
  return 'Alice';
}

function getAge(): number {
  return 30;
}

describe('Type checking', () => {
  it('should return a string', () => {
    const result = getName();
    expect(typeof result).toBe('string');
  });

  it('should return a number', () => {
    const result = getAge();
    expect(typeof result).toBe('number');
  });
});

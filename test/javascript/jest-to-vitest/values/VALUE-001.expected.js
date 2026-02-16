import { describe, it, expect } from 'vitest';

describe('greeting service', () => {
  it('should generate a personalized greeting', () => {
    const name = 'Alice';
    const greeting = `Hello, ${name}!`;
    expect(greeting).toBe(`Hello, ${name}!`);
  });

  it('should handle template literals with expressions', () => {
    const a = 5;
    const b = 10;
    expect(`Sum is ${a + b}`).toBe('Sum is 15');
  });

  it('should support nested template expressions', () => {
    const user = { first: 'Jane', last: 'Doe' };
    const display = `${user.first} ${user.last}`;
    expect(display).toBe('Jane Doe');
  });
});

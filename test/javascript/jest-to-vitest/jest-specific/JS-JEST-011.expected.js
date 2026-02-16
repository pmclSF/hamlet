import { describe, it, expect } from 'vitest';
import { transformSync } from './babel-helper';
import { parseTemplate } from './template-engine';

describe('Template engine with Babel transforms', () => {
  it('compiles JSX templates', () => {
    const input = '<div className="app">Hello</div>';
    const compiled = transformSync(input);
    expect(compiled).toContain('createElement');
  });

  it('parses template expressions', () => {
    const result = parseTemplate('Hello, {{ name }}!', { name: 'World' });
    expect(result).toBe('Hello, World!');
  });

  it('handles optional chaining transform', () => {
    const code = 'const x = obj?.foo?.bar;';
    const output = transformSync(code);
    expect(output).toBeDefined();
  });
});

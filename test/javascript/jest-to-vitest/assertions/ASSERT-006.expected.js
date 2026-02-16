import { describe, it, expect } from 'vitest';

describe('TemplateEngine', () => {
  it('should include the username in the greeting', () => {
    const output = render('greeting', { name: 'Alice' });
    expect(output).toContain('Alice');
  });

  it('should match an email pattern', () => {
    const email = generateEmail('bob');
    expect(email).toMatch(/^[^@]+@[^@]+\.[^@]+$/);
  });

  it('should match a substring in the rendered template', () => {
    const html = render('page', { title: 'Home' });
    expect(html).toMatch('Home');
  });

  it('should contain the doctype declaration', () => {
    const html = render('page', { title: 'Test' });
    expect(html).toMatch(/^<!DOCTYPE html>/i);
  });
});

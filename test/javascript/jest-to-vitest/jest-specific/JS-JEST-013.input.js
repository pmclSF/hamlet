/**
 * @jest-environment jsdom
 */

describe('DOM manipulation', () => {
  it('creates an element', () => {
    const div = document.createElement('div');
    div.textContent = 'Hello';
    document.body.appendChild(div);
    expect(document.body.innerHTML).toContain('Hello');
  });

  it('queries elements', () => {
    const span = document.createElement('span');
    span.setAttribute('data-testid', 'greeting');
    span.textContent = 'Hi there';
    document.body.appendChild(span);
    const found = document.querySelector('[data-testid="greeting"]');
    expect(found.textContent).toBe('Hi there');
  });
});

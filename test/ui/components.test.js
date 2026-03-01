import {
  bar,
  confidenceBadge,
  copyBtn,
  esc,
  frameworkBadge,
} from '../../src/ui/components.js';

describe('ui components', () => {
  it('esc should HTML-escape untrusted values', () => {
    const raw = '<script>alert("xss")</script> & "quoted"';
    const escaped = esc(raw);
    expect(escaped).toContain('&lt;script&gt;');
    expect(escaped).toContain('&quot;xss&quot;');
    expect(escaped).toContain('&amp;');
    expect(escaped).not.toContain('<script>');
  });

  it('frameworkBadge should mark unknown framework values as muted', () => {
    const html = frameworkBadge('');
    expect(html).toContain('unknown');
    expect(html).toContain('badge-muted');
  });

  it('confidenceBadge should select severity variant by threshold', () => {
    expect(confidenceBadge(90)).toContain('badge-success');
    expect(confidenceBadge(65)).toContain('badge-warning');
    expect(confidenceBadge(10)).toContain('badge-danger');
  });

  it('bar should clamp values to 0..100', () => {
    expect(bar(150, 'accent')).toContain('width:100%');
    expect(bar(-10, 'accent')).toContain('width:0%');
  });

  it('copyBtn should escape copied text payloads', () => {
    const html = copyBtn('"><img src=x onerror=alert(1)>');
    expect(html).not.toContain('<img');
    expect(html).toContain('data-copy');
  });
});

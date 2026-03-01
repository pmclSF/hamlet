import { buildSideBySide, computeDiff } from '../../src/ui/diff.js';

describe('ui diff', () => {
  it('should compute standard LCS diff for small inputs', () => {
    const diff = computeDiff('a\nb\nc', 'a\nx\nc');
    expect(diff).toEqual([
      { type: 'equal', value: 'a' },
      { type: 'delete', value: 'b' },
      { type: 'insert', value: 'x' },
      { type: 'equal', value: 'c' },
    ]);
  });

  it('should use large-input fallback without throwing', () => {
    const oldText = Array.from({ length: 2500 }, (_, i) => `old-${i}`).join(
      '\n'
    );
    const newText = Array.from({ length: 2500 }, (_, i) => `new-${i}`).join(
      '\n'
    );

    const diff = computeDiff(oldText, newText);
    expect(diff.length).toBe(5000);
    expect(diff[0]).toEqual({ type: 'delete', value: 'old-0' });
    expect(diff[1]).toEqual({ type: 'insert', value: 'new-0' });
  });

  it('should build side-by-side rows from diff entries', () => {
    const pairs = buildSideBySide([
      { type: 'equal', value: 'same' },
      { type: 'delete', value: 'gone' },
      { type: 'insert', value: 'new' },
    ]);

    expect(pairs).toEqual([
      {
        left: 'same',
        right: 'same',
        leftNum: 1,
        rightNum: 1,
        type: 'equal',
      },
      {
        left: 'gone',
        right: null,
        leftNum: 2,
        rightNum: null,
        type: 'delete',
      },
      {
        left: null,
        right: 'new',
        leftNum: null,
        rightNum: 2,
        type: 'insert',
      },
    ]);
  });
});

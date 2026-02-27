/**
 * Line-based diff using LCS (Longest Common Subsequence).
 * Returns an array of { type, value } entries.
 *
 * @param {string} oldText
 * @param {string} newText
 * @returns {Array<{type: 'equal'|'insert'|'delete', value: string}>}
 */
export function computeDiff(oldText, newText) {
  const oldLines = oldText.split('\n');
  const newLines = newText.split('\n');
  const n = oldLines.length;
  const m = newLines.length;

  // Build LCS table
  const dp = Array.from({ length: n + 1 }, () => new Uint16Array(m + 1));
  for (let i = 1; i <= n; i++) {
    for (let j = 1; j <= m; j++) {
      if (oldLines[i - 1] === newLines[j - 1]) {
        dp[i][j] = dp[i - 1][j - 1] + 1;
      } else {
        dp[i][j] = Math.max(dp[i - 1][j], dp[i][j - 1]);
      }
    }
  }

  // Backtrack to produce diff
  const result = [];
  let i = n;
  let j = m;
  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      result.push({ type: 'equal', value: oldLines[i - 1] });
      i--;
      j--;
    } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
      result.push({ type: 'insert', value: newLines[j - 1] });
      j--;
    } else {
      result.push({ type: 'delete', value: oldLines[i - 1] });
      i--;
    }
  }
  result.reverse();
  return result;
}

/**
 * Build side-by-side pairs from a diff result.
 * Each pair has { left, right, leftNum, rightNum, type }.
 */
export function buildSideBySide(diffEntries) {
  const pairs = [];
  let oldNum = 0;
  let newNum = 0;
  for (const entry of diffEntries) {
    if (entry.type === 'equal') {
      oldNum++;
      newNum++;
      pairs.push({
        left: entry.value,
        right: entry.value,
        leftNum: oldNum,
        rightNum: newNum,
        type: 'equal',
      });
    } else if (entry.type === 'delete') {
      oldNum++;
      pairs.push({
        left: entry.value,
        right: null,
        leftNum: oldNum,
        rightNum: null,
        type: 'delete',
      });
    } else {
      newNum++;
      pairs.push({
        left: null,
        right: entry.value,
        leftNum: null,
        rightNum: newNum,
        type: 'insert',
      });
    }
  }
  return pairs;
}

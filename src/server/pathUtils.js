import fs from 'fs/promises';
import path from 'path';

/**
 * Validate and resolve a user-supplied path, ensuring it stays within rootDir.
 *
 * Rejects null bytes, resolves against rootDir, and verifies the result is a
 * child of rootDir (or rootDir itself). This prevents path traversal attacks
 * where user input like "../../etc/passwd" could escape the project root.
 *
 * When the target exists on disk, both root and target are resolved through
 * fs.realpath so that symlinks pointing outside root are caught. When the
 * target does not yet exist (e.g. an output path), the lexical check still
 * applies.
 *
 * @param {string} userPath - The untrusted path from the request
 * @param {string} rootDir - The trusted project root directory (must be absolute)
 * @returns {Promise<string>} The safe, resolved absolute path
 * @throws {Error} If the path is invalid or escapes the root
 */
export async function safePath(userPath, rootDir) {
  if (typeof userPath !== 'string' || !userPath) {
    throw new Error('Path must be a non-empty string');
  }

  if (userPath.includes('\0')) {
    throw new Error('Path contains null bytes');
  }

  const resolvedRoot = path.resolve(rootDir);
  const resolved = path.resolve(resolvedRoot, userPath);

  // Lexical containment check (always applied, even when target does not exist)
  if (
    resolved !== resolvedRoot &&
    !resolved.startsWith(resolvedRoot + path.sep)
  ) {
    throw new Error('Path outside project root');
  }

  // Symlink-aware containment: resolve both through realpath when possible.
  // If the target does not exist, realpath will throw and we fall back to the
  // lexical check above (which already passed).
  try {
    const realRoot = await fs.realpath(resolvedRoot);
    const realTarget = await fs.realpath(resolved);

    if (
      realTarget !== realRoot &&
      !realTarget.startsWith(realRoot + path.sep)
    ) {
      throw new Error('Path outside project root');
    }
  } catch (err) {
    // Re-throw our own containment errors
    if (err.message === 'Path outside project root') {
      throw err;
    }
    // ENOENT — target doesn't exist yet, lexical check is sufficient
  }

  return resolved;
}

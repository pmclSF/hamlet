import path from 'path';

/**
 * Validate and resolve a user-supplied path, ensuring it stays within rootDir.
 *
 * Rejects null bytes, resolves against rootDir, and verifies the result is a
 * child of rootDir (or rootDir itself). This prevents path traversal attacks
 * where user input like "../../etc/passwd" could escape the project root.
 *
 * @param {string} userPath - The untrusted path from the request
 * @param {string} rootDir - The trusted project root directory (must be absolute)
 * @returns {string} The safe, resolved absolute path
 * @throws {Error} If the path is invalid or escapes the root
 */
export function safePath(userPath, rootDir) {
  if (typeof userPath !== 'string' || !userPath) {
    throw new Error('Path must be a non-empty string');
  }

  if (userPath.includes('\0')) {
    throw new Error('Path contains null bytes');
  }

  const resolvedRoot = path.resolve(rootDir);
  const resolved = path.resolve(resolvedRoot, userPath);

  // The resolved path must be exactly the root or start with root + separator.
  // Checking root + sep prevents prefix collisions like /app matching /application.
  if (
    resolved !== resolvedRoot &&
    !resolved.startsWith(resolvedRoot + path.sep)
  ) {
    throw new Error('Path outside project root');
  }

  return resolved;
}

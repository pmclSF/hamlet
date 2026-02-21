/**
 * IR-driven Selenium WebDriver emitter (stub).
 *
 * Selenium conversions currently use legacy converters.
 * This stub provides the unified emitter interface so Selenium
 * can participate in IR pipeline coverage reporting.
 *
 * emitFullFile returns null → pipeline falls back to ir-patch → legacy.
 */

import {
  Assertion,
  Navigation,
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  RawCode,
  Comment,
  ImportStatement,
} from '../../ir.js';

// ═══════════════════════════════════════════════════════════════════════
// Unified interface
// ═══════════════════════════════════════════════════════════════════════

export function emitNode(_node) {
  return { code: '', supported: false };
}

export function matchesBaseline(_line, _node) {
  return false;
}

/**
 * Stub — returns null so pipeline falls back to legacy emit.
 */
export function emitFullFile(_ir) {
  return null;
}

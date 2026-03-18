#!/usr/bin/env node

// terrain-compat.js — Backward-compatible shim for the `terrain` npm binary.
//
// The npm `terrain` command is deprecated in favor of `terrain-convert`.
// This shim prints a one-time deprecation notice, then delegates to the
// real converter CLI (bin/terrain.js).
//
// The Go-based `terrain` binary (installed via `go install`) is the primary
// Terrain CLI for test system intelligence. This npm binary handles only
// test framework conversion.

import { createRequire } from 'module';

const require = createRequire(import.meta.url);
const version = require('../package.json').version;

// Detect if we're being invoked as the npm `terrain` binary.
// If so, show deprecation notice on stderr (won't pollute piped output).
const binName = process.argv[1] || '';
const isNpmBin =
  binName.includes('node_modules') || binName.endsWith('terrain-compat.js');

if (isNpmBin) {
  process.stderr.write(
    `\x1b[33m` +
      `[terrain-testframework] The npm "terrain" command is deprecated.\n` +
      `Use "terrain-convert" instead for test framework conversion.\n` +
      `The "terrain" command now refers to the Go-based test intelligence CLI.\n` +
      `Install it: go install github.com/pmclSF/terrain/cmd/terrain@latest\n` +
      `\x1b[0m\n`
  );
}

// Delegate to the real converter CLI.
await import('./terrain.js');

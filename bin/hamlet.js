#!/usr/bin/env node

// DEPRECATED: The 'hamlet' command has been renamed to 'terrain'.
// This alias will be removed in a future release.
console.error(
  "WARNING: The 'hamlet' command has been renamed to 'terrain'.\n" +
    "         The 'hamlet' alias is deprecated and will be removed in a future release.\n"
);

// Re-export the terrain CLI.
import './terrain.js';

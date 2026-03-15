// Weak coverage: no tests cover this module
import { createConnection } from './connection.js';
import type { DbConfig } from './connection.js';

export async function runMigrations(config: DbConfig) {
  const conn = createConnection(config);
  await conn.query('CREATE TABLE IF NOT EXISTS users (id TEXT, email TEXT, password_hash TEXT)');
  await conn.close();
}

export async function rollbackMigration(config: DbConfig, version: number) {
  const conn = createConnection(config);
  await conn.query(`DELETE FROM migrations WHERE version = ${version}`);
  await conn.close();
}

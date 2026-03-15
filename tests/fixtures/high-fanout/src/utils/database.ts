export interface DbConnection {
  host: string;
  port: number;
  database: string;
}

export function connect(config: DbConnection): { connected: boolean; config: DbConnection } {
  return { connected: true, config };
}

export function disconnect(): boolean {
  return true;
}

export function query(sql: string, params: unknown[] = []): unknown[] {
  return [{ sql, params, rows: 0 }];
}

export function transaction<T>(fn: () => T): T {
  return fn();
}

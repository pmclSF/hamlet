// Weak coverage: no tests cover this module
export interface DbConfig {
  host: string;
  port: number;
  database: string;
}

export function createConnection(config: DbConfig) {
  return {
    query: async (sql: string) => {
      return { rows: [], rowCount: 0 };
    },
    close: async () => {},
  };
}

export function createPool(config: DbConfig, poolSize: number = 10) {
  return {
    acquire: async () => createConnection(config),
    release: async () => {},
    drain: async () => {},
  };
}

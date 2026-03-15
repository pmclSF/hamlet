// Weak coverage: imported by many modules but not directly tested
export interface AppConfig {
  sessionTtl: number;
  rateLimit: number;
  dbHost: string;
  cacheEnabled: boolean;
}

const defaultConfig: AppConfig = {
  sessionTtl: 3600,
  rateLimit: 100,
  dbHost: 'localhost',
  cacheEnabled: true,
};

export function getConfig(): AppConfig {
  return { ...defaultConfig };
}

export function mergeConfig(overrides: Partial<AppConfig>): AppConfig {
  return { ...defaultConfig, ...overrides };
}

// Weak coverage: no tests import this module directly
export function log(level: string, message: string): void {
  console.log(`[${level}] ${message}`);
}

export function error(message: string): void {
  log('ERROR', message);
}

export function warn(message: string): void {
  log('WARN', message);
}

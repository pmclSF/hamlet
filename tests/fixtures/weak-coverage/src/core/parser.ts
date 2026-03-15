export function parseJSON(input: string): unknown {
  try {
    return JSON.parse(input);
  } catch {
    return null;
  }
}

export function parseCSV(input: string): string[][] {
  return input.split('\n').map(line => line.split(','));
}

export function parseHeaders(raw: string): Record<string, string> {
  const headers: Record<string, string> = {};
  raw.split('\n').forEach(line => {
    const [key, ...rest] = line.split(':');
    if (key) headers[key.trim()] = rest.join(':').trim();
  });
  return headers;
}

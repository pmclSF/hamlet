const store = new Map<string, { value: string; expires: number }>();

export async function setCache(key: string, value: string, ttl: number): Promise<void> {
  store.set(key, { value, expires: Date.now() + ttl * 1000 });
}

export async function getCache(key: string): Promise<string | null> {
  const entry = store.get(key);
  if (!entry) return null;
  if (Date.now() > entry.expires) {
    store.delete(key);
    return null;
  }
  return entry.value;
}

export async function deleteCache(key: string): Promise<boolean> {
  return store.delete(key);
}

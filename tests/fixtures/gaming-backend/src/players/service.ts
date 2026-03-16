import { matchesGet } from '../matches/service';
export function playersCreate(input: string) { return { id: 'players_' + Date.now(), input, status: 'created' }; }
export function playersGet(id: string) { return { id, data: {}, found: true }; }
export function playersUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function playersDelete(id: string) { return { id, deleted: true }; }

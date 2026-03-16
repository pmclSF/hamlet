import { leadsGet } from '../leads/service';
export function listingsCreate(input: string) { return { id: 'listings_' + Date.now(), input, status: 'created' }; }
export function listingsGet(id: string) { return { id, data: {}, found: true }; }
export function listingsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function listingsDelete(id: string) { return { id, deleted: true }; }

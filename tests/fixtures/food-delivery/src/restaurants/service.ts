import { menusGet } from '../menus/service';
export function restaurantsCreate(input: string) { return { id: 'restaurants_' + Date.now(), input, status: 'created' }; }
export function restaurantsGet(id: string) { return { id, data: {}, found: true }; }
export function restaurantsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function restaurantsDelete(id: string) { return { id, deleted: true }; }

import { enrollmentGet } from '../enrollment/service';
export function coursesCreate(input: string) { return { id: 'courses_' + Date.now(), input, status: 'created' }; }
export function coursesGet(id: string) { return { id, data: {}, found: true }; }
export function coursesUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function coursesDelete(id: string) { return { id, deleted: true }; }

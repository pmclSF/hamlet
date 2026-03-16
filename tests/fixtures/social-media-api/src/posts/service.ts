import { commentsGet } from '../comments/service';
export function postsCreate(input: string) { return { id: 'posts_' + Date.now(), input, status: 'created' }; }
export function postsGet(id: string) { return { id, data: {}, found: true }; }
export function postsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function postsDelete(id: string) { return { id, deleted: true }; }

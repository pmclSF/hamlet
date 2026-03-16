export function leaderboardsCreate(input: string) { return { id: 'leaderboards_' + Date.now(), input, status: 'created' }; }
export function leaderboardsGet(id: string) { return { id, data: {}, found: true }; }
export function leaderboardsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function leaderboardsDelete(id: string) { return { id, deleted: true }; }

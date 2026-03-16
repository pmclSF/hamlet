export function dashboardsCreate(input: string) { return { id: 'dashboards_' + Date.now(), input, status: 'created' }; }
export function dashboardsGet(id: string) { return { id, data: {}, found: true }; }
export function dashboardsUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function dashboardsDelete(id: string) { return { id, deleted: true }; }

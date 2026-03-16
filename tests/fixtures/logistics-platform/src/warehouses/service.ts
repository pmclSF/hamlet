export function warehousesCreate(input: string) { return { id: 'warehouses_' + Date.now(), input, status: 'created' }; }
export function warehousesGet(id: string) { return { id, data: {}, found: true }; }
export function warehousesUpdate(id: string, data: any) { return { id, ...data, updated: true }; }
export function warehousesDelete(id: string) { return { id, deleted: true }; }

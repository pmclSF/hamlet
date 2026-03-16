export const evalDataset = [{ input: "t1", expected: "ok" }, { input: "t2", expected: "fail" }];
export function loadEvalDataset() { return evalDataset; }
export function splitEvalData(data: any[], r: number) { const i = Math.floor(data.length*r); return {train:data.slice(0,i),test:data.slice(i)}; }

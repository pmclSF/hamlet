import { analyzerAction } from '../risk/analyzer';
export function placeOrder(symbol: string, qty: number) { return { orderId: 'ord_' + Date.now(), symbol, qty, status: 'placed' }; }
export function cancelOrder(orderId: string) { return { orderId, status: 'cancelled' }; }
export function fillOrder(orderId: string) { return { orderId, status: 'filled' }; }

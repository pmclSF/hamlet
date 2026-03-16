export function initiatePurchase(productId: string, platform: string) {
  return { purchaseId: 'pur_' + Date.now(), productId, platform, status: 'initiated' };
}

export function completePurchase(purchaseId: string) {
  return { purchaseId, status: 'completed' };
}

export function validateReceipt(receiptData: string, platform: string) {
  return { valid: true, platform };
}

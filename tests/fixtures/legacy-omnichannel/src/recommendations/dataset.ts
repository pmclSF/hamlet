export const productDataset = [
  { id: 'prod_1', name: 'Laptop', category: 'electronics' },
  { id: 'prod_2', name: 'Shoes', category: 'apparel' },
];

export function loadProductDataset() { return productDataset; }

export function loadUserBehavior(userId: string) {
  return [{ productId: 'prod_1', action: 'view' }];
}

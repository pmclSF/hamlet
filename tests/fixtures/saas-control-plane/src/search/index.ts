export function indexDocument(collection: string, doc: any) {
  return { indexed: true, collection, docId: doc.id || 'doc_1' };
}

export function searchDocuments(collection: string, query: string) {
  return { collection, query, results: [], total: 0 };
}

export function deleteIndex(collection: string) {
  return { collection, deleted: true };
}

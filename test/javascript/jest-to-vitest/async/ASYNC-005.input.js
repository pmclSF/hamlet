describe('Database tests', () => {
  let db;

  beforeEach(async () => {
    db = await connectToDatabase();
    await db.clear();
  });

  afterEach(async () => {
    await db.disconnect();
  });

  it('inserts a record', async () => {
    await db.insert({ id: 1, name: 'Alice' });
    const record = await db.findById(1);
    expect(record.name).toBe('Alice');
  });

  it('deletes a record', async () => {
    await db.insert({ id: 2, name: 'Bob' });
    await db.delete(2);
    const record = await db.findById(2);
    expect(record).toBeNull();
  });
});

describe('SQL query builder', () => {
  it('should build a multiline query string', () => {
    const table = 'users';
    const query = `
      SELECT id, name, email
      FROM ${table}
      WHERE active = true
      ORDER BY name ASC
    `;
    expect(query).toContain('SELECT id, name, email');
    expect(query).toContain(`FROM ${table}`);
    expect(query).toContain('ORDER BY name ASC');
  });

  it('should produce correct HTML template', () => {
    const title = 'Hello';
    const html = `<div>
  <h1>${title}</h1>
  <p>Welcome to the site</p>
</div>`;
    expect(html).toContain('<h1>Hello</h1>');
    expect(html).toContain('<p>Welcome to the site</p>');
  });
});

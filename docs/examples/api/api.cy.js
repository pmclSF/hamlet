describe('API Tests', () => {
    it('should make a successful API call', () => {
      cy.request('GET', '/api/users').then((response) => {
        expect(response.status).to.eq(200);
        expect(response.body).to.have.length.greaterThan(0);
      });
    });
  });
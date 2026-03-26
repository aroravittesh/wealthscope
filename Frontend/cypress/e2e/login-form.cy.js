describe('Login page', () => {
  it('lets user fill the login form', () => {
    cy.visit('/auth/login');

    cy.get('input[name="email"]').type('test@example.com');
    cy.get('input[name="password"]').type('123456');

    // Button should become enabled once ngForm becomes valid
    cy.get('button[type="submit"]').should('not.be.disabled');

    cy.get('input[name="email"]').should('have.value', 'test@example.com');
  });
});


/**
 * Admin dashboard (Sprint 3). Uses a non-expired JWT-shaped token and ADMIN user in
 * localStorage so guards pass; admin APIs are stubbed so no backend is required.
 */
describe('Admin dashboard', () => {
  const futureExp = Math.floor(Date.now() / 1000) + 60 * 60 * 24 * 365;
  // Must be standard base64 (not base64url) so AuthService's atob() can decode exp/role.
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const payload = btoa(JSON.stringify({ exp: futureExp, role: 'ADMIN' }));
  const adminToken = `${header}.${payload}.sig`;

  beforeEach(() => {
    cy.intercept('GET', '**/api/admin/users', []).as('getUsers');
    cy.intercept('GET', '**/api/admin/assets', []).as('getAssets');
  });

  it('shows the dashboard and switches to the Assets tab', () => {
    cy.visit('/admin', {
      onBeforeLoad(win) {
        win.localStorage.setItem('authToken', adminToken);
        win.localStorage.setItem(
          'user',
          JSON.stringify({
            email: 'admin@test.com',
            riskPreference: 'MEDIUM',
            role: 'ADMIN',
          })
        );
      },
    });

    cy.contains('h2', 'Admin Dashboard').should('be.visible');
    cy.wait('@getUsers');
    cy.wait('@getAssets');

    cy.contains('button', 'Assets').click();
    cy.contains('h3', 'Create Asset').should('be.visible');
    cy.get('input[name="symbol"]').type('TEST');
    cy.get('input[name="symbol"]').should('have.value', 'TEST');
  });
});

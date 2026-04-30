describe('WealthScope E2E Flow', () => {
  it('should load the landing page and navigate to login when Get Started is clicked', () => {
    // Visit the frontend root
    cy.visit('/');
    
    // Verify landing page content
    cy.contains('WealthScope').should('be.visible');
    
    // Attempt interaction logic
    cy.contains('Get Started').click();
    
    // Verify routing 
    cy.url().should('include', '/auth/login');
    
    // Verify target page rendered
    cy.contains('Sign In').should('be.visible');
  });
});

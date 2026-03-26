import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { AuthService } from './auth.service';

describe('AuthService', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [AuthService]
    });
    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
    
    // Clear storage to isolate tests
    localStorage.clear();
  });

  afterEach(() => {
    httpMock.verify();
  });

  // 1. register
  it('should register a new user successfully (#register)', (done) => {
    service.register('new@test.com', 'pass', 'New User').subscribe(res => {
      expect(res.user.email).toBe('new@test.com');
      done();
    });
  });

  // 2. login
  it('should login an existing user (#login)', (done) => {
    service.login('test@example.com', 'password123').subscribe(res => {
      expect(res.user.email).toBe('test@example.com');
      expect(localStorage.getItem('authToken')).toBeTruthy();
      done();
    });
  });

  // 3. logout
  it('should clear local storage & state on logout (#logout)', () => {
    localStorage.setItem('authToken', 'fake_token');
    service.logout();
    expect(localStorage.getItem('authToken')).toBeNull();
    expect(service.getCurrentUser()).toBeNull();
  });

  // 4. refreshToken
  it('should fail refreshToken if none exists (#refreshToken)', (done) => {
    localStorage.removeItem('refreshToken');
    service.refreshToken().subscribe({
      error: (err) => {
        expect(err.error.message).toBe('No refresh token available');
        done();
      }
    });
  });

  // 5. getCurrentUser
  it('should return null if no user is logged in (#getCurrentUser)', () => {
    expect(service.getCurrentUser()).toBeNull();
  });

  // 6. isAuthenticated
  it('should return authentication status (#isAuthenticated)', () => {
    expect(service.isAuthenticated()).toBeFalse();
  });

  // 7. handleAuthResponse (private method tested via public)
  it('should appropriately handle auth responses internally (#handleAuthResponse)', (done) => {
    service.login('demo@finsight.com', 'Demo@123').subscribe(() => {
      expect(localStorage.getItem('user')).toBeTruthy();
      done();
    });
  });

  // 8. hasToken (private method tested implicitly via initial state)
  it('should derive token presence (#hasToken)', () => {
    expect(service['hasToken']()).toBeFalse();
  });

  // 9. getUserFromStorage (private method tested explicitly)
  it('should parse user from storage correctly (#getUserFromStorage)', () => {
    localStorage.setItem('user', JSON.stringify({ id: '99', email: 'stored@test' }));
    const u = service['getUserFromStorage']();
    expect(u?.id).toEqual('99');
  });

  // 10. isTokenExpired (private method test)
  it('should evaluate token expiry mechanics properly (#isTokenExpired)', () => {
    expect(service['isTokenExpired']()).toBeTrue();
  });

  // 11. checkTokenExpiry (private method called in constructor)
  it('should init expiry checker logic safely (#checkTokenExpiry)', () => {
    // Constructor handles this call, we assert it doesn't break
    expect(service).toBeTruthy();
  });

  // 12. generateMockToken (private string generator)
  it('should generate a valid JWT format mock token (#generateMockToken)', () => {
    const token = service['generateMockToken']({ id: '1', email: 'e', fullName: 'n', role: 'USER', createdAt: new Date(), updatedAt: new Date() });
    expect(token.split('.').length).toBe(3);
  });
});

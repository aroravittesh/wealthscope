import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from 'rxjs';

import { AuthService } from '../../../services/auth.service';
import { SignupComponent } from './signup';

describe('Signup', () => {
  let component: SignupComponent;
  let fixture: ComponentFixture<SignupComponent>;

  const authServiceMock = {
    register: jasmine.createSpy('register').and.returnValue(of(void 0)),
    // Not used by this spec, but included to satisfy any runtime access.
    isAuthenticated$: of(false),
    getCurrentUser: () => null,
    logout: () => {},
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SignupComponent, RouterTestingModule],
      providers: [{ provide: AuthService, useValue: authServiceMock }],
    })
    .compileComponents();

    fixture = TestBed.createComponent(SignupComponent);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

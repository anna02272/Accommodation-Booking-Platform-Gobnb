import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PasswordStrenghtValidatorComponent } from './password-strenght-validator.component';

describe('PasswordStrenghtValidatorComponent', () => {
  let component: PasswordStrenghtValidatorComponent;
  let fixture: ComponentFixture<PasswordStrenghtValidatorComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ PasswordStrenghtValidatorComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(PasswordStrenghtValidatorComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});

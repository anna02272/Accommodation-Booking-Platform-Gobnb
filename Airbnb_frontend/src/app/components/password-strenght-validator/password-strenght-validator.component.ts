import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';

@Component({
  selector: 'app-password-strenght-validator',
  templateUrl: './password-strenght-validator.component.html',
  styleUrls: ['./password-strenght-validator.component.css']
})
export class PasswordStrenghtValidatorComponent implements OnChanges {
  @Input() password: string = '';
  strength: string = ''; 

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['password']) {
      this.updatePasswordStrength();
    }
  }

  updatePasswordStrength() {
    const passw = this.password;

    const moderate = /(?=.*[A-Z])(?=.*[a-z]).{5,}|(?=.*[\d])(?=.*[a-z]).{5,}|(?=.*[\d])(?=.*[A-Z])(?=.*[a-z]).{5,}/g;
    const strong = /(?=.*[A-Z])(?=.*[a-z])(?=.*[\d]).{7,}|(?=.*[\!@#$%^&*()\\[\]{}\-_+=~`|:;"'<>,./?])(?=.*[a-z])(?=.*[\d]).{7,}/g;
    const extraStrong = /(?=.*[A-Z])(?=.*[a-z])(?=.*[\d])(?=.*[\!@#$%^&*()\\[\]{}\-_+=~`|:;"'<>,./?]).{9,}/g;

    
    if (this.testPasswRegexp(passw, extraStrong)) {
      this.strength = 'extra';
    } else if (this.testPasswRegexp(passw, strong)) {
      this.strength = 'strong';
    } else if (this.testPasswRegexp(passw, moderate)) {
      this.strength = 'moderate';
    } else if (passw.length > 0) {
      this.strength = 'weak';
    } else if (passw.length == 0) {
      this.strength = '';
    }
  }

  testPasswRegexp(password: string, regexp: RegExp): boolean {
    return regexp.test(password);
  }
}

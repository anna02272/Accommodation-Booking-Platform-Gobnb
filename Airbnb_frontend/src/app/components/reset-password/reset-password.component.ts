import { Component, OnInit } from '@angular/core';
import {  ActivatedRoute, Router } from '@angular/router';
import { AuthService } from 'src/app/services';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

@Component({
  selector: 'app-reset-password',
  templateUrl: './reset-password.component.html',
  styleUrls: ['./reset-password.component.css']
})
export class ResetPasswordComponent implements OnInit{
  notification: any = { msgType: '', msgBody: '' };
  email!: string
  passwordForm!: FormGroup;

  constructor(
    private authService: AuthService, 
    private router: Router,
    private route: ActivatedRoute,
    private formBuilder: FormBuilder
    ) {
  }
  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      this.email = params['email'];
    });
    this.passwordForm = this.formBuilder.group({
      passwordResetToken: ['', Validators.required],
      password: ['', [Validators.required, this.passwordValidator()]],
      passwordConfirm: ['', Validators.required],
    });
  }

  resetPassword() {
    if (this.passwordForm.invalid) {
      this.notification = { msgType: 'error', msgBody: 'Please fix the form errors.' };
      return;
    }

    const { password, passwordConfirm, passwordResetToken } = this.passwordForm.value;
    if (password !== passwordConfirm) {
      this.notification = { msgType: 'error', msgBody: 'Password and password confirm do not match.' };
      return;
    }

    this.authService.resetPassword(passwordResetToken, password, passwordConfirm).subscribe(
      () => {
        this.notification = { msgType: 'success', msgBody: 'Password reset successful.' };
       console.log('Password reset successful');
        this.router.navigate(['/login']);
      },
      (error) => {
        this.notification = { msgType: 'error', msgBody: 'Password reset failed.' };
        console.error('Password reset failed', error);
      }
    );
  }
  passwordValidator() {
    return Validators.pattern(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/);
  }
}

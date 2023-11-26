import { Component, OnInit } from '@angular/core';
import {  ActivatedRoute, Router } from '@angular/router';
import { AuthService } from 'src/app/services';
import { Validators } from '@angular/forms';

@Component({
  selector: 'app-reset-password',
  templateUrl: './reset-password.component.html',
  styleUrls: ['./reset-password.component.css']
})
export class ResetPasswordComponent implements OnInit{
  notification: any = { msgType: '', msgBody: '' };
  email!: string
  passwordResetToken: string = '';
  password: string = '';
  passwordConfirm: string = '';

  constructor(
    private authService: AuthService, 
    private router: Router,
    private route: ActivatedRoute
    ) {
  }
  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      this.email = params['email'];
    });
   
  }

  resetPassword() {
    if (this.password !== this.passwordConfirm) {
      this.notification = { msgType: 'error', msgBody: 'Password and password confirm do not match.' };
      return;
    }

    this.authService.resetPassword(this.passwordResetToken, this.password, this.passwordConfirm).subscribe(
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

import { Component } from '@angular/core';
import { AuthService } from 'src/app/services';
import { Router } from '@angular/router';

@Component({
  selector: 'app-forgot-password',
  templateUrl: './forgot-password.component.html',
  styleUrls: ['./forgot-password.component.css']
})

export class ForgotPasswordComponent {
  email: string = '';
  notification: any = { msgType: '', msgBody: '' };

  constructor(private authService: AuthService,
    private router: Router) {}

  onSubmit() {
    this.authService.forgotPassword(this.email).subscribe(
      () => {
        this.notification = { msgType: 'success', msgBody: 'Password reset email sent. Check your inbox.' };
        this.router.navigate(['/reset-password'], { queryParams: { email: this.email }});
      },
      (error) => {
        this.notification = { msgType: 'error', msgBody: 'Failed to send password reset email. Please try again.' };
        console.error('Forgot password error', error);
      }
    );
  }
}

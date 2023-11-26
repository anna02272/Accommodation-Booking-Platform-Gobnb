import { Component, OnInit } from '@angular/core';
import {  ActivatedRoute, Router } from '@angular/router';
import { AuthService } from 'src/app/services';


@Component({
  selector: 'app-email-verification',
  templateUrl: './email-verification.component.html',
  styleUrls: ['./email-verification.component.css']
})
export class EmailVerificationComponent implements OnInit {
  email!: string
  verificationCode: string = '';
  notification = { msgType: '', msgBody: '' };

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

  verifyEmail() {
      this.authService.verifyEmail(this.verificationCode) 
        .subscribe(
          () => {
            this.notification = { msgType: 'success', msgBody: `Email verification success.` };
            this.router.navigate(['/login']);
            console.log('Email verification success');
          },
          error => {
            this.notification = { msgType: 'error', msgBody: `Email verification failed.` };
            console.error('Email verification failed', error);
          }
        );
  }
  resendVerification() {
    if (this.email) {
      this.authService.resendVerificationEmail(this.email)
        .subscribe(
          () => {
            this.notification = { msgType: 'success', msgBody: `Verification email resent successfully.` };
            console.log('Verification email resent successfully');
          },
          error => {
            this.notification = { msgType: 'success', msgBody: `Failed to resend verification email.` };
            console.error('Failed to resend verification email', error);
          }
        );
    } else {
      console.error('Email not provided for resending verification');
    }
  }
}

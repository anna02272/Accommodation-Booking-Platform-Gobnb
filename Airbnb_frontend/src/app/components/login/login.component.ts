import { Component, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router, Params } from '@angular/router';
import { AuthService, UserService } from '../../services';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { NgZone } from '@angular/core';



interface DisplayMessage {
  msgType: string;
  msgBody: string;
}

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit, OnDestroy {
  form: FormGroup = new FormGroup({});
  submitted = false;
  notification: DisplayMessage = {} as DisplayMessage;
  returnUrl = '';
  private ngUnsubscribe: Subject<void> = new Subject<void>();
  recaptchaSiteKey = "6Lcm8hwpAAAAAK-MQIOvQQNNUdTPNzjI2PCZMVKs";


  constructor(
    private userService: UserService,
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute,
    private formBuilder: FormBuilder,
    private zone: NgZone 

  ) {}

  ngOnInit() {
    this.route.params
      .pipe(takeUntil(this.ngUnsubscribe))
      .subscribe((params: Params) => {
        this.notification = params as DisplayMessage || { msgType: '', msgBody: '' };
      });
  

    this.returnUrl = this.route.snapshot.queryParams['returnUrl'] || '/';

    this.form = this.formBuilder.group({
      email: [
        '',
        Validators.compose([
          Validators.required,
          Validators.minLength(1),
          Validators.maxLength(32)
        ]) 
      ],
      password: [
        '',
        Validators.compose([
          Validators.required,
          Validators.minLength(8),
          Validators.maxLength(32)
        ])
      ],
      captcha: [null, Validators.required]
    });
  }

  ngOnDestroy() {
    this.ngUnsubscribe.next();
    this.ngUnsubscribe.complete();
  }

   //recaptcha
   handleSuccess(event: any): void {
    console.log('reCAPTCHA success:', event);
  }

  handleReset(): void {
    console.log('reCAPTCHA reset');
  }

  handleExpire(): void {
    console.log('reCAPTCHA expired');
  }

  handleLoad(): void {
    console.log('reCAPTCHA loaded');
  }


onSubmit() {
  this.notification = { msgType: '', msgBody: '' };
  this.submitted = true;

  if (this.form.get('captcha')?.invalid && this.form.get('captcha')?.untouched) {
    this.notification = {
      msgType: 'error',
      msgBody: 'Please check the reCAPTCHA.'
    };
    this.submitted = false; 
    return;
  }

  this.authService.login(this.form.value).subscribe(
    () => {
      this.userService.getMyInfo().subscribe();
      this.router.navigate(['/home']);
    },
    (error) => {
      console.log("error")
      console.log(error)
      this.submitted = false;
      
      if (error.statusText === 'Unknown Error') {
        this.notification = {
          msgType: 'error',
          msgBody: 'Authorization service not available.'
        };
      } else {
        this.notification = {
          msgType: 'error',
          msgBody: 'Incorrect username or password.'
        };
      }
    }
  );
}




}
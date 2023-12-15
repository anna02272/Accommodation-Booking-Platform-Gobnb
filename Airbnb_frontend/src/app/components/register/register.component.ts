import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators ,AbstractControl} from '@angular/forms';
import { ActivatedRoute, Router, Params } from '@angular/router';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { AuthService } from 'src/app/services';

interface DisplayMessage {
  msgType: string;
  msgBody: string;
}

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
  password: string = '';
  personalInfoForm: FormGroup = new FormGroup({});
  submitted = false;
  address: FormGroup = new FormGroup({});
  street: any;
  name : any;
  new : any;


  notification: DisplayMessage = {} as DisplayMessage;
  returnUrl = '';
  private ngUnsubscribe: Subject<void> = new Subject<void>();
  recaptchaSiteKey = "6Lcm8hwpAAAAAK-MQIOvQQNNUdTPNzjI2PCZMVKs";

  constructor(
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute,
    private formBuilder: FormBuilder
  ) {
    
    this.personalInfoForm = this.formBuilder.group({
      username: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(32)]],
      password: ['', [Validators.required, Validators.minLength(8), Validators.maxLength(32)]],
      email: ['', [Validators.required, Validators.email, Validators.minLength(6), Validators.maxLength(64)]],
      name: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(64)]],
      lastname: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(64)]],
      
            street: ['', Validators.required],
            city: ['', Validators.required],
            country: ['', Validators.required],
      
      
      age: ['', [Validators.maxLength(3)]],
      gender: [''],
      userRole: ['', [Validators.required]],
    });

  }
  ngOnInit() {
    this.route.params
      .pipe(takeUntil(this.ngUnsubscribe))
      .subscribe((params: Params) => {
        this.notification = params as DisplayMessage || { msgType: '', msgBody: '' };
      });

      const passwordPatternValidator = (control: AbstractControl): { [key: string]: boolean } | null => {
        const passwordPattern = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/;
        const valid = passwordPattern.test(control.value);
        return valid ? null : { 'invalidPassword': true };
      };

    this.returnUrl = this.route.snapshot.queryParams['returnUrl'] || '/';


  }
  get passwordControl() {
    return this.personalInfoForm.get('password');
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
    this.street= this.personalInfoForm.get('street')
    this.name= this.personalInfoForm.get('name')
    const emailControl = this.personalInfoForm.get('email');
    
    this.new = {}
    this.new.username= this.personalInfoForm.get('username')?.value
    this.new.password= this.personalInfoForm.get('password')?.value
    this.new.email= this.personalInfoForm.get('email')?.value
    this.new.name= this.personalInfoForm.get('name')?.value
    this.new.lastname= this.personalInfoForm.get('lastname')?.value
    this.new.address= {}
    this.new.address.street= this.personalInfoForm.get('street')?.value
    this.new.address.city= this.personalInfoForm.get('city')?.value
    this.new.address.country= this.personalInfoForm.get('country')?.value
    this.new.age= this.personalInfoForm.get('age')?.value
    this.new.gender= this.personalInfoForm.get('gender')?.value
    this.new.userRole= this.personalInfoForm.get('userRole')?.value

    if (this.personalInfoForm.get('captcha')?.invalid && this.personalInfoForm.get('captcha')?.untouched) {
    this.notification.msgBody += 'Please check the reCAPTCHA. ';
    this.submitted = false;
    console.log("captcha failed")
    return
    }
    this.authService.register(this.new).subscribe(
      (data) => {
        console.log("register")
        this.submitted = true;
        const email = emailControl?.value;
          this.notification = { msgType: 'success', msgBody: `You are registered! Check your email (${email}) for verification.` };
          this.router.navigate(['/email-verification'],  { queryParams: { email: email }});
      
    },
      (error) => {
              console.log(this.personalInfoForm.value)

        // console.error('Registration error', error);
        // this.notification = { msgType: 'error', msgBody: 'Registration failed. Please try again./Username alredy exists' };
        // this.submitted = false;
        // // Handle  error
      if (error.status === 409) {
        if (error.error.message === 'Username already exists') {
          this.notification = { msgType: 'error', msgBody: 'Registration failed. Username already exists' };
        } else if (error.error.message === 'Email already exists') {
          this.notification = { msgType: 'error', msgBody: 'Registration failed. Email already exists' };
        } else {
          this.notification = { msgType: 'error', msgBody: 'Registration failed. Please try again.' };
        }
      } 
    
      else if (error.status === 400) {
        this.notification = { msgType: 'error', msgBody: 'Password is in blacklist. Use other password.' };
       }

       else {
        this.notification = { msgType: 'error', msgBody: 'Registration failed. Please try again.' };

  //         if (this.personalInfoForm.get('captcha')?.invalid && this.personalInfoForm.get('captcha')?.untouched) {
  //        this.notification.msgBody += 'Please check the reCAPTCHA. ';
  // }  
  
      }


    
      this.submitted = false;
      }
    );
    

  }

  }


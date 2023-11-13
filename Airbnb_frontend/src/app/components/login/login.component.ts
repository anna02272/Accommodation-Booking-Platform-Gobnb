import { Component, OnDestroy, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router, Params } from '@angular/router';
import { AuthService, UserService } from '../../services';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

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

  constructor(
    private userService: UserService,
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute,
    private formBuilder: FormBuilder
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
          Validators.minLength(4),
          Validators.maxLength(64)
        ])
      ],
      password: [
        '',
        Validators.compose([
          Validators.required,
          Validators.minLength(4),
          Validators.maxLength(32)
        ])
      ]
    });
  }

  ngOnDestroy() {
    this.ngUnsubscribe.next();
    this.ngUnsubscribe.complete();
  }

  onSubmit() {
    this.notification = { msgType: '', msgBody: '' };
    this.submitted = true;
  
    this.authService.login(this.form.value).subscribe(
      () => {
        this.userService.getMyInfo().subscribe();
        this.router.navigate(["/home"]);
      },
      (error) => {
        this.submitted = false;
        this.notification = {
          msgType: 'error',
          msgBody: 'Incorrect username or password.'
        };
      }
    );
  }
  
}

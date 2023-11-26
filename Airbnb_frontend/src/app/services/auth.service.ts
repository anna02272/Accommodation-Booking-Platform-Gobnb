import { Injectable } from '@angular/core';
import { HttpHeaders } from '@angular/common/http';
import { ApiService } from './api.service';
import { UserService } from './user.service';
import { ConfigService } from './config.service';
import { Router } from '@angular/router';
import { map } from 'rxjs/operators';



@Injectable()
export class AuthService {

  constructor(
    private apiService: ApiService,
    private userService: UserService,
    private config: ConfigService,
    private router: Router
  ) {
  }

  private access_token = null;

  login(user: any) {
    const loginHeaders = new HttpHeaders({
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    });

    const body = {
      'email': user.email,
      'password': user.password
    };
    return this.apiService.post(this.config.login_url, JSON.stringify(body), loginHeaders)
      .pipe(map((res) => {
        console.log('Login success');
        this.access_token = res.accessToken;
        localStorage.setItem("jwt", res.accessToken)

        return this.userService.getMyInfo();
      }));
  }

  register(user: any) {
    const signupHeaders = new HttpHeaders({
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    });
    const body = {
      'username': user.username,
      'password': user.password,
      'email' : user.email,
      'name' : user.name,
      'lastname' : user.lastname,
      'address' : {  
        'street' : user.address?.street,
        'city' : user.address?.city,
        'country' : user.address?.country,
      },
      'age' : user.age,
      'gender' : user.gender,
      'userRole' : user.userRole,
     
    };
    return this.apiService.post(this.config.register_url, JSON.stringify(body), signupHeaders)
      .pipe(map(() => {
        console.log('Register success');
      }));
  }

  verifyEmail(verificationCode: string) {
    return this.apiService.get(this.config.verifyEmail_url + `/${verificationCode}`);
  }

  resendVerificationEmail(email: string) {
    return this.apiService.get(this.config.resendVerification_url + `/${email}`);
  }

  forgotPassword(email: string) {
    const body = {
      'email': email
    };
    return this.apiService.post(this.config.forgotPassword_url, JSON.stringify(body));
  }

  resetPassword(passwordResetToken: string, password: string, passwordConfirm: string) {
    const body = {
      'passwordResetToken' : passwordResetToken,
      'password': password,
      'passwordConfirm': passwordConfirm
    };
    return this.apiService.patch(this.config.resetPassword_url + `/${passwordResetToken}`, JSON.stringify(body));
  }

  logout() {
    this.userService.currentUser = null;
    this.access_token = null;
    this.router.navigate(['/login']);
  }

  tokenIsPresent() {
    return this.access_token != undefined && this.access_token != null;
  }

  getToken() {
    return this.access_token;
  }
  getRole() {
    return this.userService.currentUser.user.userRole;
  }
 
}

import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { MatDatepickerModule } from '@angular/material/datepicker';
import { MatInputModule } from '@angular/material/input';
import { MatNativeDateModule } from '@angular/material/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {MatFormFieldModule} from '@angular/material/form-field';



import { AppComponent } from './app.component';
import { LoginComponent } from './components/login/login.component';
import { RegisterComponent } from './components/register/register.component';
import { HomeComponent } from './components/home/home.component';
import { HeaderComponent } from './components/header/header.component';
import { AppRoutingModule } from './app-routing.module';
import { AccommodationsComponent } from './components/accommodations/accommodations.component';
import { AccommodationComponent } from './components/accommodation/accommodation.component';
import { ReservationComponent } from './components/reservation/reservation.component';
import { ReservationsComponent } from './components/reservations/reservations.component';
import { ProfileComponent } from './components/profile/profile.component';
import { SearchComponent } from './components/search/search.component';
import { EditProfileComponent } from './components/edit-profile/edit-profile.component';
import { CreateAccommodationComponent } from './components/create-accommodation/create-accommodation.component';
import { EmailVerificationComponent } from './components/email-verification/email-verification.component';
import { MatDialogModule } from '@angular/material/dialog';
import { ApiService, AuthService, ConfigService, UserService } from './services';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { TokenInterceptor } from './interceptor/TokenInterceptor';
import { ReactiveFormsModule } from '@angular/forms';
import { FormsModule } from '@angular/forms';
import { PasswordStrenghtValidatorComponent } from './components/password-strenght-validator/password-strenght-validator.component';
import { ReservationService } from './services/reservation.service';
import { CommonModule, DatePipe } from '@angular/common';
import { ForgotPasswordComponent } from './components/forgot-password/forgot-password.component';
import { ResetPasswordComponent } from './components/reset-password/reset-password.component';
import { NgxCaptchaModule } from 'ngx-captcha';
import { AccommodationService } from './services/accommodation.service';
import { RefreshService } from './services/refresh.service';
import { DeleteAccountComponent } from './components/delete-account/delete-account.component';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    RegisterComponent,
    HeaderComponent,
    HomeComponent,
    AccommodationsComponent,
    AccommodationComponent,
    ReservationComponent,
    ReservationsComponent,
    ProfileComponent,
    SearchComponent,
    EditProfileComponent,
    CreateAccommodationComponent,
    EmailVerificationComponent,
    PasswordStrenghtValidatorComponent,
    ForgotPasswordComponent,
    ResetPasswordComponent,
    DeleteAccountComponent,
  ],
  imports: [
    HttpClientModule,
    BrowserModule,
    AppRoutingModule,
    MatDatepickerModule,
    MatInputModule,
    MatNativeDateModule,
    BrowserAnimationsModule,
    MatFormFieldModule,
    MatDialogModule,
    ReactiveFormsModule,
    FormsModule,
    CommonModule,
    DatePipe,
    NgxCaptchaModule
    // RecaptchaModule,
    // RecaptchaFormsModule

  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TokenInterceptor,
      multi: true,
    },

  //    {
  //   provide: RECAPTCHA_SETTINGS,
  //   useValue: {
  //     siteKey: '6Lcm8hwpAAAAAK-MQIOvQQNNUdTPNzjI2PCZMVKs',
  //   } as RecaptchaSettings,
  // },
    ConfigService,
    ApiService,
    AuthService,
    UserService,
    ReservationService,
    DatePipe,
    AccommodationService,
    RefreshService
    
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }

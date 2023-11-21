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
import { MobileVerificationComponent } from './components/mobile-verification/mobile-verification.component';
import { MatDialogModule } from '@angular/material/dialog';
import { ApiService, AuthService, ConfigService, UserService } from './services';
import { HTTP_INTERCEPTORS, HttpClientModule } from '@angular/common/http';
import { TokenInterceptor } from './interceptor/TokenInterceptor';
import { ReactiveFormsModule } from '@angular/forms';
import { FormsModule } from '@angular/forms';
import { PasswordStrenghtValidatorComponent } from './components/password-strenght-validator/password-strenght-validator.component';
import { ReservationService } from './services/reservation.service';
import { CommonModule, DatePipe } from '@angular/common';


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
    MobileVerificationComponent,
    PasswordStrenghtValidatorComponent,
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
    DatePipe

  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TokenInterceptor,
      multi: true,
    },
    ConfigService,
    ApiService,
    AuthService,
    UserService,
    ReservationService,
    DatePipe
    
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }

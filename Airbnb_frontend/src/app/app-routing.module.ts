import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { LoginComponent } from './components/login/login.component';
import { RegisterComponent } from './components/register/register.component';
import { HomeComponent } from './components/home/home.component';
import { AccommodationComponent } from './components/accommodation/accommodation.component';
import { ProfileComponent } from './components/profile/profile.component';
import { EditProfileComponent } from './components/edit-profile/edit-profile.component';
import { CreateAccommodationComponent } from './components/create-accommodation/create-accommodation.component';
import { AuthGuard } from './services/auth.guard';
import { EmailVerificationComponent } from './components/email-verification/email-verification.component';
import { ForgotPasswordComponent } from './components/forgot-password/forgot-password.component';
import { ResetPasswordComponent } from './components/reset-password/reset-password.component';
import { ReservationsComponent } from './components/reservations/reservations.component';

const routes: Routes = [
  {
      path: 'login',
      component: LoginComponent
    },
  {
    path: 'register',
    component: RegisterComponent
  },
  {
    path: 'home',
    component: HomeComponent
  },
  {
    path: 'accommodation',
    component: AccommodationComponent
  },
  { path: 'accommodation/:id',
  component: AccommodationComponent  },
  {
    path: 'profile',
    component: ProfileComponent,
    canActivate: [AuthGuard] 
  },
  {
    path: 'edit-profile',
    component: EditProfileComponent,
    canActivate: [AuthGuard] 
  },
  {
    path: 'create-accommodation',
    component: CreateAccommodationComponent,
    canActivate: [AuthGuard] ,
    data: {
      roles: ['Host']
    }
  },
  {
    path: 'email-verification',
    component: EmailVerificationComponent ,
  },
  {
    path: 'forgot-password',
    component: ForgotPasswordComponent ,
  },
  {
    path: 'reset-password',
    component: ResetPasswordComponent ,
  },
  {
    path: 'reservations',
    component: ReservationsComponent,
    canActivate: [AuthGuard] ,
    data: {
      roles: ['Guest']
    }
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }

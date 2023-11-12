import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { LoginComponent } from './components/login/login.component';
import { RegisterComponent } from './components/register/register.component';
import { HomeComponent } from './components/home/home.component';
import { AccommodationComponent } from './components/accommodation/accommodation.component';
import { ProfileComponent } from './components/profile/profile.component';
import { EditProfileComponent } from './components/edit-profile/edit-profile.component';
import { ReservationComponent } from './components/reservation/reservation.component';
import { MobileVerificationComponent } from './components/mobile-verification/mobile-verification.component';
import { CreateAccommodationComponent } from './components/create-accommodation/create-accommodation.component';


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
  {
    path: 'profile',
    component: ProfileComponent
  },
  {
    path: 'edit-profile',
    component: EditProfileComponent
  },
  {
    path: 'verification',
    component: MobileVerificationComponent
  },
  {
    path: 'create-accommodation',
    component: CreateAccommodationComponent
  },


];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }

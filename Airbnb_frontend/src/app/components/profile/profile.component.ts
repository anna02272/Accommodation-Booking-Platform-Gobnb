import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService, UserService } from 'src/app/services';
import { User } from 'src/app/models/user';
import { HttpClient } from '@angular/common/http';
import { concatMap } from 'rxjs/operators';


@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileComponent {
  errorMessage: string | null = null;
  currentProfile!: User;
  notifications!: any[];
  notifServiceAvailable: boolean = false;
  profileServiceAvailable: boolean = false;
  hostEmail!: string;
  hostIdP!: string;
  hostFeatured!: boolean;



  constructor(private userService: UserService, private router: Router, private authService: AuthService, private httpClient: HttpClient) {
    
  }
  ngOnInit() {
    this.load();

    
  }
  load() {
    this.userService.getProfile().subscribe((data: any) => {
      this.currentProfile = data.user;
      console.log(this.currentProfile.name)
      console.log(this.currentProfile.lastname)

      console.log(this.currentProfile.gender)
      console.log(this.currentProfile.address.city)

      this.getNotifications();

      this.hostEmail = this.currentProfile.email;

      this.httpClient.get('https://localhost:8000/api/profile/getUser/' + this.hostEmail).pipe(
        concatMap((response: any) => {
          this.hostIdP = response.user.id;
          //alert("hostIdP " + this.hostIdP);
          return this.httpClient.get('https://localhost:8000/api/profile/isFeatured/' + this.hostIdP);
        })
      ).subscribe(
        (response: any) => {
          this.hostFeatured = response.featured;
          //alert("hostFeatured " + this.hostFeatured);
        },
        error => {
          console.error('Error', error);
        }
      );


  },

  error => {
    console.log("here")
    console.log(error)
      if (error.statusText === 'Unknown Error') {
        console.log("here if unknown error")
          this.profileServiceAvailable = true;
      }
  }
  
  );
}

isFeatured(){
  return this.hostFeatured;
}

deleteProfile() {
    this.userService.deleteProfile().subscribe(
      () => {
        console.log('Profile deleted successfully')
        console.log("here")
        this.errorMessage = null;
        this.authService.logout();
        this.router.navigate(['/delete-account']);
      },
      error => {
        console.error('Failed to delete profile:', error);
          console.log('Error object:', error.error); // Log the entire error object


        if (error.status === 400 && error.error.message) {
          this.errorMessage = error.error.message;
        } else {
          this.errorMessage = "Failed to delete profile. Please try again.";
        }
      }
    );
  }

  getNotifications() {
    this.userService.getUserNotifications().subscribe(
      (data: any) => {
        this.notifications = data;
        console.log("notifications")
        console.log(data)
      },
      error => {
         if (error.statusText === 'Unknown Error') {
          this.notifServiceAvailable = true;
       
      }
        // console.error('Failed to delete profile:', error);
          console.log('Error object:', error.error); // Log the entire error object

      }
    );
  }
  
  getRole() {
    return this.userService.currentUser?.user.userRole;
  }
}

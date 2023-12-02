import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { AuthService, UserService } from 'src/app/services';

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileComponent {
  errorMessage: string | null = null;

  constructor(private userService: UserService, private router: Router, private authService: AuthService) {
    
  }

deleteProfile() {
  this.userService.deleteProfile().subscribe(
    () => {
      console.log('Profile deleted successfully')
      console.log("here")
      this.errorMessage = null;
      this.authService.logout();
      this.router.navigate(['/register']);

    },
    error => {
      console.error('Failed to delete profile:', error);
      this.errorMessage = error; 
    }
  );
}

}

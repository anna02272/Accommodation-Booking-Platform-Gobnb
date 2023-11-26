import { Component } from '@angular/core';
import { UserService } from 'src/app/services';

@Component({
  selector: 'app-edit-profile',
  templateUrl: './edit-profile.component.html',
  styleUrls: ['./edit-profile.component.css']
})
export class EditProfileComponent {
  currentPassword = '';
  newPassword = '';
  confirmNewPassword = '';
  currentUsername = '';

  constructor(private userService: UserService) {}

  changePassword() {
    if (this.newPassword !== this.confirmNewPassword) {
      alert('New passwords do not match.');
      return;
    }
    const user = this.userService.currentUser;
    console.log("1",this.currentPassword, this.newPassword, this.confirmNewPassword)


    this.userService.changePassword(this.currentPassword, this.newPassword, this.confirmNewPassword)
      .subscribe(
       () => {
          alert('Password changed successfully.');
        },
        (_:any) => {
          alert('Failed to change password.');
        }
      );
  }
}

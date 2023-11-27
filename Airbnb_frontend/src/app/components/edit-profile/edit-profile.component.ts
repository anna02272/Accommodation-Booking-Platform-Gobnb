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
  notification = { msgType: '', msgBody: '' };

  constructor(private userService: UserService) {}

  changePassword() {
    if (this.newPassword !== this.confirmNewPassword) {
      this.notification = { msgType: 'error', msgBody: `New passwords do not match.` };
      return;
    }
    const user = this.userService.currentUser;
    console.log("1",this.currentPassword, this.newPassword, this.confirmNewPassword)


    this.userService.changePassword(this.currentPassword, this.newPassword, this.confirmNewPassword)
      .subscribe(
       () => {
        this.notification = { msgType: 'success', msgBody: `Password changed successfully.` };
        this.resetForm();
        },
        (_:any) => {
          this.notification = { msgType: 'error', msgBody: `Failed to change password.` };
        }
      );
      
  }
  resetForm() {
    this.currentPassword = '';
    this.newPassword = '';
    this.confirmNewPassword = '';
    this.currentUsername = '';
  }
}

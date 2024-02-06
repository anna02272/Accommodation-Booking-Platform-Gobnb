import { Component, OnInit } from '@angular/core';
import { UserService } from 'src/app/services';
import { User } from 'src/app/models/user';
import { HttpClient } from '@angular/common/http';
import { ElementRef, ViewChild } from '@angular/core';
import { FormBuilder, FormGroup, Validators ,AbstractControl} from '@angular/forms';





@Component({
  selector: 'app-edit-profile',
  templateUrl: './edit-profile.component.html',
  styleUrls: ['./edit-profile.component.css']
})
export class EditProfileComponent implements OnInit {
  currentUser: User = {} as User; 
  currentPassword = '';
  newPassword = '';
  confirmNewPassword = '';
  currentUsername ='';
  currentEmail= "";
  address : FormGroup = new FormGroup({});
  new : any
  currentRole='';
  currentName= '';
  changeInfoForm: FormGroup = new FormGroup({});
  currentProfile!: User;
 profileServiceAvailable: boolean = false;
  notification = { msgType: '', msgBody: '' };
  notification2 = { msgType: '', msgBody: '' };

  constructor(private userService: UserService, private http: HttpClient,private formBuilder: FormBuilder
    ) {
    this.currentEmail=this.userService.currentUser.user.email
    this.currentUsername=this.userService.currentUser.user.username
    this.currentRole=this.userService.currentUser.user.userRole
    // console.log(this.currentProfile.name)
    this.currentName=this.userService.currentUser.user.name

    this.changeInfoForm = this.formBuilder.group({
   
      email: ['', [Validators.required, Validators.email, Validators.minLength(6), Validators.maxLength(64)]],
      name: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(64)]],
      lastname: ['', [Validators.required, Validators.minLength(1), Validators.maxLength(64)]],
            street: ['', Validators.required],
            city: ['', Validators.required],
            country: ['', Validators.required],
      age: ['', [Validators.maxLength(3)]],
      gender: [''],
    });
   
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
  },
   (error) => {
    if (error.statusText === 'Unknown Error') {
      this.profileServiceAvailable = true;
      }
  } 
  
  );
}
  // getName() {
  //   console.log(this.userService.currentUserProfile.user.name)

  //   return this.userService.currentUserProfile.user.name
  // }

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
  saveChanges(){
    
    this.new = {}
    this.new.username= this.currentUsername
    this.new.email= this.currentEmail
    this.new.name= this.changeInfoForm.get('name')?.value
    this.new.lastname= this.changeInfoForm.get('lastname')?.value
    this.new.address= {}
    this.new.address.street= this.changeInfoForm.get('street')?.value
    this.new.address.city= this.changeInfoForm.get('city')?.value
    this.new.address.country= this.changeInfoForm.get('country')?.value
    this.new.age= this.changeInfoForm.get('age')?.value
    this.new.gender= this.changeInfoForm.get('gender')?.value
    this.new.userRole= this.currentRole

    console.log(this.new)


    this.http.post('https://localhost:8000/api/profile/updateUser', this.new).subscribe(
      () => {
        this.notification2 = { msgType: 'success', msgBody: 'Profile updated successfully.' };
      },
      (error) => {
        console.error('Error updating profile', error);
        this.notification2 = { msgType: 'error', msgBody: 'Failed to update profile.' };

         if (error.statusText === 'Unknown Error') {
           this.profileServiceAvailable = true;
       }

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

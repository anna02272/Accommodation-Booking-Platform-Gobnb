import { Component } from '@angular/core';
import { UserService } from 'src/app/services';


@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.css']

})
export class HomeComponent {
  constructor( 
    private userService: UserService
    ) 
  { }

  getRole() {
    return this.userService.currentUser.user.userRole;
  }
}

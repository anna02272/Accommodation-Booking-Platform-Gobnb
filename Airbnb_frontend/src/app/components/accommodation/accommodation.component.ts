import { Component } from '@angular/core';
import { UserService } from 'src/app/services';

@Component({
  selector: 'app-accommodation',
  templateUrl: './accommodation.component.html',
  styleUrls: ['./accommodation.component.css']
})
export class AccommodationComponent {
  constructor( 
    private userService: UserService
    ) 
  { }

  getRole() {
    return this.userService.currentUser?.user.userRole;
  }
}

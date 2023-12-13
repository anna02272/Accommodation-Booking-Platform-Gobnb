import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Accommodation } from 'src/app/models/accommodation';
import { UserService } from 'src/app/services';
import { AccommodationService } from 'src/app/services/accommodation.service';

@Component({
  selector: 'app-accommodation',
  templateUrl: './accommodation.component.html',
  styleUrls: ['./accommodation.component.css']
})
export class AccommodationComponent implements OnInit {
  accId!: string; 
  hostId!: string;
  accommodation!: Accommodation;
  //am_map!: Map<string, boolean>;
  
  constructor( 
    private userService: UserService,
    private accService : AccommodationService,
    private route: ActivatedRoute ,
    ) 
  { }
 

  ngOnInit(): void {
    this.accId = this.route.snapshot.paramMap.get('id')!;
    this.accService.getById(this.accId).subscribe((accommodation: Accommodation) => {
      this.accommodation = accommodation;
      this.hostId = accommodation.host_id;
    });
    //this.am_map = this.accommodation.accommodation_amenities;
  }

  getRole() {
    return this.userService.currentUser?.user.userRole;
  }
}

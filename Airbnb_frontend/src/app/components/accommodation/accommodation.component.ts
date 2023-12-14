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
  tv!: boolean;
  wifi!: boolean;
  ac!: boolean;
  am_map!: Map<string, boolean>;
  
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
      // generate an empty map string:booelan
      this.am_map = new Map<string, boolean>();
      //this am_map becomes accommodation_amenities
      //this.am_map = JSON.stringify(this.accommodation.accommodation_amenities);
      this.am_map = Object.entries(this.accommodation.accommodation_amenities).reduce((map, [key, value]) => map.set(key, value), new Map<string, boolean>());
      //console.log(this.am_map.get('TV'));
      this.tv = this.am_map.get('TV')!;
      this.wifi = this.am_map.get('WiFi')!;
      this.ac = this.am_map.get('AC')!;
    });
    //this.am_map = this.accommodation.accommodation_amenities;
  }

  getRole() {
    return this.userService.currentUser?.user.userRole;
  }
}

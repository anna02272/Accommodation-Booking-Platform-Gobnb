import { Component, OnInit } from '@angular/core';
import { RateHost } from 'src/app/models/rateHost';
import { RateAccommodation } from 'src/app/models/rateAccommodation';

import { RatingService } from 'src/app/services/rating.service';
import { RefreshService } from 'src/app/services/refresh.service';

@Component({
  selector: 'app-ratings',
  templateUrl: './ratings.component.html',
  styleUrls: ['./ratings.component.css']
})
export class RatingsComponent implements OnInit{
  ratingResponse!: RateHost;
  ratingResponse2!: RateAccommodation;
  notification = { msgType: '', msgBody: '' };
  ratingServiceAvailable: boolean = false;
  
  constructor(
    private ratingService: RatingService,
    private refreshService: RefreshService,
  ) {}
  ngOnInit() {
    this.load();
    this.subscribeToRefresh();
  }
  load() {
    this.ratingService.getAll().subscribe((data: RateHost) => {
      this.ratingResponse = data;
      
    
  },
   (error) => {
    if (error.statusText === 'Unknown Error') {
       console.log("here")
       console.log(error)
      this.ratingServiceAvailable = true;
      }
  }
  );
  this.ratingService.getAllAccommodation().subscribe((data2: RateAccommodation) => {
    this.ratingResponse2 = data2;
});
  }
  
  private subscribeToRefresh() {
    this.refreshService.getRefreshObservable().subscribe(() => {
      this.load();
    });
  }
  }
import { Component, OnInit } from '@angular/core';
import { RateHost } from 'src/app/models/rateHost';
import { RatingService } from 'src/app/services/rating.service';
import { RefreshService } from 'src/app/services/refresh.service';

@Component({
  selector: 'app-ratings',
  templateUrl: './ratings.component.html',
  styleUrls: ['./ratings.component.css']
})
export class RatingsComponent implements OnInit{
  ratingResponse!: RateHost;
  notification = { msgType: '', msgBody: '' };
  
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
  });
  }
  private subscribeToRefresh() {
    this.refreshService.getRefreshObservable().subscribe(() => {
      this.load();
    });
  }
  }
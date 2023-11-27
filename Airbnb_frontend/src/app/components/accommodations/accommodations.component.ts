import { Component, OnInit } from '@angular/core';
import { Accommodation } from 'src/app/models/accommodation';
import { UserService } from 'src/app/services';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { RefreshService } from 'src/app/services/refresh.service';

@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit{
  accommodations: Accommodation[] = [];

  constructor(
    private accService: AccommodationService,
    private refreshService: RefreshService,
  ) {}

  ngOnInit() {
    this.load();
    this.subscribeToRefresh();
  
  }

  load() {
      this.accService.getAll().subscribe((data: Accommodation[]) => {
        this.accommodations = data;
  });
}
private subscribeToRefresh() {
  this.refreshService.getRefreshObservable().subscribe(() => {
    this.load();
  });
}

}

import { Component, OnInit } from '@angular/core';
import { GetReservation } from 'src/app/models/GetReservation';
import { RefreshService } from 'src/app/services/refresh.service';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-reservations',
  templateUrl: './reservations.component.html',
  styleUrls: ['./reservations.component.css']
})
export class ReservationsComponent implements OnInit{
reservations: GetReservation[] = [];
notification = { msgType: '', msgBody: '' };
reservationServiceAvailable: boolean = false;

constructor(
  private resService: ReservationService,
  private refreshService: RefreshService,
) {}
ngOnInit() {
  this.load();
  this.subscribeToRefresh();
}
load() {
  this.resService.getAll().subscribe((data: GetReservation[]) => {
    this.reservations = data;
},
 error => {
if (error.statusText === 'Unknown Error') {
       console.log("here")
       console.log(error)
      this.reservationServiceAvailable = true;
      }
 }
);
}

private subscribeToRefresh() {
  this.refreshService.getRefreshObservable().subscribe(() => {
    this.load();
  });
}
cancelReservation(id: string): void {
  this.resService.cancelReservation(id).subscribe(
    () => {
        this.refreshService.refresh();
        this.notification = { msgType: 'success', msgBody: `Reservation canceled successfully.` };
    },
  error => {
    if (error.status === 400 && error.error && error.error.error) {
      const errorMessage = error.error.error;
      this.notification = { msgType: 'error', msgBody: errorMessage };
    } else {
      this.notification = { msgType: 'error', msgBody: 'An error occurred while processing your request.' };
    }
  }
);
}
}

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
});
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
    (error) => {
      this.notification = { msgType: 'error', msgBody: `Cannot cancel reservation, check-in date has already started` };
      console.error('Error canceling reservation:', error);
    }
  );
}
}

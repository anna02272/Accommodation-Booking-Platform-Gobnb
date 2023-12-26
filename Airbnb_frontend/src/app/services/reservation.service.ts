import { ConfigService } from "./config.service";
import { Observable } from "rxjs";
import { Reservation } from "../models/reservation";
import { Injectable } from "@angular/core";
import { ApiService } from "./api.service";
import { map } from 'rxjs/operators';

@Injectable()
export class ReservationService {
  reservationCreated!: Reservation;
  constructor(
    private configService: ConfigService,     
    private apiService: ApiService
) {}

  createReservation(reservation: Reservation): Observable<any> {
     return this.apiService.post(this.configService.createReservation_url, reservation)
      .pipe(map((reservationCreatedDb: Reservation) => {
        this.reservationCreated = reservationCreatedDb;
        return reservationCreatedDb;
      }));
  }
  getAll() {
    return this.apiService.get(this.configService.resv_url + "/getAll");
   }
   
   cancelReservation(id: string): Observable<void> {
    return this.apiService.delete(`${this.configService.resv_url}/cancel/${id}`);
  }

     checkAvailability(checkAvailabilityData: any, accId: string) {
    return this.apiService.post(this.configService.resv_url + "/availability/" + accId, checkAvailabilityData);
   }

   checkPrice(checkAvailabilityData: any, accId: string) {
    return this.apiService.post(this.configService.resv_url + "/prices/" + accId, checkAvailabilityData);
   }
}


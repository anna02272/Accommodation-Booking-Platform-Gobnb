import { HttpClient } from "@angular/common/http";
import { ConfigService } from "./config.service";
import { Observable } from "rxjs";
import { Reservation } from "../models/reservation";
import { Injectable } from "@angular/core";
import { ApiService } from "./api.service";
import { map } from 'rxjs/operators';

@Injectable()
export class ReservationService {
  reservationCreated!: Reservation;
  constructor(private http: HttpClient, private configService: ConfigService,     
    private apiService: ApiService
) {}

  createReservation(reservation: Reservation): Observable<any> {
     return this.apiService.post(this.configService.createReservation_url, reservation)
      .pipe(map((reservationCreatedDb: Reservation) => {
        this.reservationCreated = reservationCreatedDb;
        return reservationCreatedDb;
      }));
  }


}


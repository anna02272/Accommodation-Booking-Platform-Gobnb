import { ConfigService } from "./config.service";
import { Observable } from "rxjs";
//import { Availability } from "../models/availability";
import { AvailabilityPeriod } from "../models/availability_period";
import { Injectable } from "@angular/core";
import { ApiService } from "./api.service";
import { map } from 'rxjs/operators';
import { HttpClient, HttpHeaders } from "@angular/common/http";

@Injectable()
export class AvailabilityService {
  availabilityPeriod!: AvailabilityPeriod;
  private apiUrl = 'https://localhost:8082/api/availability/create';

  constructor(
    private configService: ConfigService,     
    private apiService: ApiService,
    private http: HttpClient
) {}

  createAvailabilityPeriod(availability_period: AvailabilityPeriod): Observable<any> {
     return this.apiService.post(this.configService.createAvailabilityPeriod_url, availability_period)
      .pipe(map((availabilityPeriodDb: AvailabilityPeriod) => {
        this.availabilityPeriod = availabilityPeriodDb;
        return availabilityPeriodDb;
      }));
  }
  // getAvailabilityByAccommodationId(id: string): Observable<void> {
  //   return this.apiService.get(`${this.configService.availability_url}/get/${id}`);
  //  }

   getAvailabilityByAccommodationId(id: string) {
    return this.apiService.get(`${this.configService.getAvailabilityPeriod_url}/${id}`);
  }
  

   sendData(data: any, accId: any): Observable<any> {
    const headers = new HttpHeaders({
      'Content-Type': 'application/json',
    });

    return this.http.post<any>(`${this.apiUrl}/${accId}`, data, { headers });
  }
//    cancelReservation(id: string): Observable<void> {
//     return this.apiService.delete(`${this.configService.resv_url}/cancel/${id}`);
//   }

//      checkAvailability(checkAvailabilityData: any, accId: string) {
//     return this.apiService.post(this.configService.resv_url + "/availability/" + accId, checkAvailabilityData);
//    }
}


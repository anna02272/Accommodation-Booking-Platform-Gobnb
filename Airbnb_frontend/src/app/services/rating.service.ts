import {Injectable} from '@angular/core';
import {ApiService} from './api.service';
import {ConfigService} from './config.service';
import { Observable } from 'rxjs';


@Injectable({
  providedIn: 'root'
})
export class RatingService {
  constructor(
    private apiService: ApiService,
    private configService: ConfigService,
  ) {
  }
  getAll() {
    return this.apiService.get(this.configService.rating_url + "/getAll");
   }
   getAllAccommodation() {
    return this.apiService.get(this.configService.rating_url + "/getAllAccomodation");
   }

   rateHost(hostId: string, rating: number): Observable<any> {
    const url = `${this.configService.rating_url}/rateHost/${hostId}`;
    const body = { rating };

    return this.apiService.post(url, body);
  }
  rateAccommodation(accommodationId: string, rating: number): Observable<any> {
    const url = `${this.configService.rating_url}/rateAccommodation/${accommodationId}`;
    console.log(url)
    const body = { rating };

    return this.apiService.post(url, body);
  }

  deleteRating(hostId: string): Observable<any> {
    const url = `${this.configService.rating_url}/deleteRating/${hostId}`;

    return this.apiService.delete(url);
  }
  deleteRatingAccommodation(accommodationId: string): Observable<any> {
    const url = `${this.configService.rating_url}/deleteRatingAccommodation/${accommodationId}`;

    return this.apiService.delete(url);
  }

  getByHostAndGuest(hostId: string): Observable<any> {
    const url = `${this.configService.rating_url}/get/${hostId}`;
    return this.apiService.get(url);
  }
  getByAccommodationAndGuest(accommodationId: string): Observable<any> {
    const url = `${this.configService.rating_url}/getAccommodation/${accommodationId}`;
    return this.apiService.get(url);
  }
  
}

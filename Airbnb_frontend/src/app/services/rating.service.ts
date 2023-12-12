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

   rateHost(hostId: string, rating: number): Observable<any> {
    const url = `${this.configService.rating_url}/rateHost/${hostId}`;
    const body = { rating };

    return this.apiService.post(url, body);
  }

  deleteRating(hostId: string): Observable<any> {
    const url = `${this.configService.rating_url}/deleteRating/${hostId}`;

    return this.apiService.delete(url);
  }

  getByHostAndGuest(hostId: string): Observable<any> {
    const url = `${this.configService.rating_url}/get/${hostId}`;
    return this.apiService.get(url);
  }
  
}

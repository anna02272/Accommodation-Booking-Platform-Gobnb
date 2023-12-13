import {Injectable} from '@angular/core';
import { ApiService } from './api.service';
import { ConfigService } from './config.service';
@Injectable()
export class AccommodationService {

  constructor(
    private apiService: ApiService,
    private config: ConfigService,
  ) {
  }
  getAll() {
    return this.apiService.get(this.config.acc_url + "/get");
   }
   getById(id : string) {
    return this.apiService.get(this.config.acc_url + "/get/" + id );
   }
   
   getByHost(hostId: string) {
    return this.apiService.get(this.config.getAccommodationsByHost_url + hostId);
   }

   deleteAccommodation(accId: string) {
    return this.apiService.delete(this.config.deleteAccommodation_url + accId)
   }

}

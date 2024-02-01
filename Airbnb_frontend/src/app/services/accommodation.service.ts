import {Injectable} from '@angular/core';
import { ApiService } from './api.service';
import { ConfigService } from './config.service';
import { HttpHeaders, HttpParams } from '@angular/common/http';
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
   getAllRecommendation(id:string) {
    return this.apiService.get(this.config.rating_url + "/getRecomendation/"+id);
   }

  getSearch(location: any, guests: any, start_date: any, end_date: any, tv: any, wifi: any, ac: any, min_price: any, max_price: any) {
    //alert(this.config.acc_url + `/get?location=${location}&guests=${guests}&start_date=${start_date}&end_date=${end_date}`);
    return this.apiService.get(this.config.acc_url + `/get?location=${location}&guests=${guests}&start_date=${start_date}&end_date=${end_date}&tv=${tv}&wifi=${wifi}&ac=${ac}&min_price=${min_price}&max_price=${max_price}`);
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

   uploadAccImages(accId: string, formData: any) {
        const boundary = '<calculated when request is sent>';

    const headers = new HttpHeaders();

    const contentTypeHeader = `multipart/form-data; boundary=${boundary}`;
    headers.set('Content-Type', contentTypeHeader);
    let returnObject = this.apiService.post(this.config.imagesUpload_url + accId, formData, headers)
    console.log(returnObject)
    return returnObject
   }

   fetchAccImages(accId: string) {
    return this.apiService.get(this.config.imagesFetch_url + accId);
   }

}

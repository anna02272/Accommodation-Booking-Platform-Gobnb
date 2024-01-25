import { ConfigService } from "./config.service";
import { Injectable } from "@angular/core";
import { ApiService } from "./api.service";

@Injectable()
export class ReportService {
  constructor(
    private configService: ConfigService,     
    private apiService: ApiService
) {}


  generateDailyReport(accId: string) {
    return this.apiService.post(this.configService.dailyReport_url + accId, {});
   }
   

   generateMonthlyReport(accId: string) {
    return this.apiService.post(this.configService.montlyReport_url + accId, {});
   }

}


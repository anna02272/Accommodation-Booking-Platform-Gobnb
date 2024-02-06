import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ReportService } from 'src/app/services/report.service';
import { UserService } from 'src/app/services/user.service';

@Component({
  selector: 'app-report',
  templateUrl: './report.component.html',
  styleUrls: ['./report.component.css']
})
export class ReportComponent implements OnInit {
   accId!: string; 
   hostId!: string;
   dailyReport!: any;
   monthlyReport!: any;
   serviceAvailable: boolean = false;

   constructor( 
    private userService: UserService,
    private reportService : ReportService,
    private route: ActivatedRoute 
    ) 

    {}

  ngOnInit(): void {
     this.accId = this.route.snapshot.paramMap.get('id')!;

     this.reportService.generateDailyReport(this.accId).subscribe((data :any) => {
      this.dailyReport = data;
      console.log(data);
      
    }, 
    (error) => {
     if (error.statusText === 'Unknown Error') {
      this.serviceAvailable = true;
      }
  });

     this.reportService.generateMonthlyReport(this.accId).subscribe((data :any) => {
      this.monthlyReport = data;
      console.log(data);
    },
     (error) => {
     if (error.statusText === 'Unknown Error') {
      this.serviceAvailable = true;
      }
  });

}
}
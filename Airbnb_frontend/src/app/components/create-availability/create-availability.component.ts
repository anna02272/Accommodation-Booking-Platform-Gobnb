import { Component, Input, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, NgForm } from '@angular/forms';
import { AvailabilityService } from 'src/app/services/availability-service';
import { AvailabilityPeriod } from 'src/app/models/availability_period';
import { DatePipe } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { UserService } from 'src/app/services';

@Component({
  selector: 'app-create-availability',
  templateUrl: './create-availability.component.html',
  styleUrls: ['./create-availability.component.css']
})
export class CreateAvailabilityComponent {
  //@Input() accId!: string;
  notification = { msgType: '', msgBody: '' };
  //get accId from url
  accId = this.route.snapshot.paramMap.get('accId');


  constructor(private dataService: AvailabilityService,
    private router: Router,
    private userService: UserService,
    private datePipe: DatePipe,
    private route: ActivatedRoute
  ) {}

  //on submit, create availability period
  onSubmit() {
    const startDate = (document.getElementById('startDate') as HTMLInputElement).value;
    //if startdate isn't set handle error
    if(startDate == ""){
      this.notification = { msgType: 'error', msgBody: `Start date is required` };
      return;
    }
    const endDate = (document.getElementById('endDate') as HTMLInputElement).value;
    //if enddate isn't set handle error
    if(endDate == ""){
      this.notification = { msgType: 'error', msgBody: `End date is required` };
      return;
    }
    //if enddate is before startdate handle error
    if(endDate < startDate){
      this.notification = { msgType: 'error', msgBody: `End date must be after start date` };
      return;
    }
    //if startdate is before today handle error
    // today is date as string in format yyyy-MM-dd, not using datepipe because it returns date as object
    var today = new Date().toISOString().slice(0,10);
    if(startDate < today.toString()){
      this.notification = { msgType: 'error', msgBody: `Start date must be after today` };
      return;
    }
    const startDateFormated = startDate +"T00:00:00Z"
    const endDateFormated = endDate + "T00:00:00Z"
    const price = parseInt((document.getElementById('price') as HTMLInputElement).value, 10);
    const priceString = (document.getElementById('price') as HTMLInputElement).value;
    if(priceString == ""){
      this.notification = { msgType: 'error', msgBody: `Price is required` };
      return;
    }
    const priceType = (document.getElementById('priceType') as HTMLInputElement).value;
    const availabilityType = (document.getElementById('availabilityType') as HTMLInputElement).value;

    const availabilityPeriodData = {
      start_date: startDateFormated,
      end_date: endDateFormated,
      price: price,
      price_type: priceType,
      availability_type: availabilityType
    };


    this.dataService.sendData(availabilityPeriodData, this.accId).subscribe(
      (response:any) => {
        this.notification = { msgType: 'success', msgBody: `Successfully created availability period;` };
        console.log('Response from server:', response);
        this.router.navigate([`/accommodation/${this.accId}`]);
      },
      (error:any) => {
        var err1 = JSON.stringify(error, Object.getOwnPropertyNames(error))
        //get error from error response
        err1 = JSON.parse(err1).error;
        this.notification = { msgType: 'error', msgBody: `Creating availability period failed: ${err1}` };
        console.error('Error:', error);
      }
    );

    //this.resetForm();
  }

  getUsername() {
    return this.userService.currentUser.user.username;
  }

}

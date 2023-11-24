import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, NgForm } from '@angular/forms';
import { ReservationService } from 'src/app/services/reservation.service';
import { Reservation } from 'src/app/models/reservation';
import { DatePipe } from '@angular/common';


@Component({
  selector: 'app-reservation',
  templateUrl: './reservation.component.html',
  styleUrls: ['./reservation.component.css']
})
export class ReservationComponent implements OnInit  {

    form!: FormGroup;
    accommodation_id!: "22151fcf-8aee-11ee-818a-0242ac120007";
    showDiv: boolean = false;
    showDivSuccess: boolean = false;
    //showDivErrorCheckInTime: boolean = false;
    check_in_date?: string;
    check_out_date?: string;
    check_in_time?: number;
    errorCheck: boolean =  false; 


    errorMessage?: "";
    successMessage?: "Reserved successfully!"
    errorMessage2?: "Please enter your check in time!"
    
    constructor(private fb: FormBuilder,private reservationService: ReservationService, 
       private datePipe: DatePipe) {

        this.showDivSuccess = false;
    }

  ngOnInit(): void {
  this.form = this.fb.group({
      check_in_time: [''],
      check_out_date: [''],
      check_in_date: ['']

    })   }


  convertToISOFormat(dateObject?: string): any {
    // Use DatePipe to format the selected date to ISO format
     if (this.datePipe.transform(dateObject, 'yyyy-MM-ddTHH:mm:ssZ') === null){
      return ""
     }
     return this.datePipe.transform(dateObject, 'yyyy-MM-ddTHH:mm:ssZ')
  }
  createReservation(): void {

    if (this.check_in_time === undefined){
      this.errorCheck = true;
      return
    } 
    else {
      if (this.check_in_time > 24 || this.check_in_time < 1){
              this.errorCheck = true;
              return
      }
    }
    const reservationCreate: Reservation = {
      accommodation_id: "22151fcf-8aee-11ee-818a-0242ac120007",
      check_in_date: this.check_in_date+`T${this.check_in_time}:00:00Z`,
      check_out_date: this.check_out_date+"T15:00:00Z"
    };

  this.reservationService.createReservation(reservationCreate).subscribe(
    {
      next: (response) => {
        console.log('Reservation created successfully', response);
        this.showDivSuccess = true;
         setTimeout(() => {
        this.showDivSuccess = false;
      }, 5000);

      },
      error: (error) => {
           this.showDiv = true;
           this.errorMessage = error.error.error;
            setTimeout(() => {
        this.showDiv = false;
      }, 5000);
           
      }
    }
  );
    
  }
}
// function ViewChild(arg0: string, arg1:
//    { static: boolean; }): (target: ReservationComponent, propertyKey: "reservationForm") => void {
//   throw new Error('Function not implemented.');
// }

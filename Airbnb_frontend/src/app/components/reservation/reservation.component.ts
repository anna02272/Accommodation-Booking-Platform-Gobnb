import { Component, Input, OnInit } from '@angular/core';
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
  @Input() accommodationId!: string; 
    form!: FormGroup;
    showDiv: boolean = false;
    showDivSuccess: boolean = false;
    showDivSuccessAvailability: boolean = false;
    //showDivErrorCheckInTime: boolean = false;
    check_in_date?: string;
    check_out_date?: string;
    check_in_time?: number;
    number_of_guests?: number;
    errorCheck: boolean =  false; 
    errorCheckGuests: boolean = false;


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
      check_in_date: [''],
      number_of_guests: ['']

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

    if (this.number_of_guests === undefined){
      this.errorCheckGuests = true;
      return
    } 
    else {
      if (this.number_of_guests < 1){
              this.errorCheckGuests = true;
              return
      }
    }

    this.errorCheckGuests = false;
    this.errorCheck = false;

    const reservationCreate: Reservation = {
      accommodation_id: this.accommodationId,
      check_in_date: this.check_in_date+`T${this.check_in_time}:00:00Z`,
      check_out_date: this.check_out_date+"T15:00:00Z",
      number_of_guests: this.number_of_guests
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
        console.log(reservationCreate)
           this.showDiv = true;
           this.errorMessage = error.error.error;
            setTimeout(() => {
        this.showDiv = false;
      }, 5000);
           
      }
    }
  );
    
  }


  
 checkAvailability(): void {

    // if (this.check_in_time === undefined){
    //   this.errorCheck = true;
    //   return
    // } 
    // else {
    //   if (this.check_in_time > 24 || this.check_in_time < 1){
    //           this.errorCheck = true;
    //           return
    //   }
    // }

    this.errorCheck = false;

    const checkAvailabilityData = {
      check_in_date: this.check_in_date+ "T00:00:00Z",
      check_out_date: this.check_out_date+"T15:00:00Z",
    };

  this.reservationService.checkAvailability(checkAvailabilityData, this.accommodationId).subscribe(
    {
      next: (response) => {
        console.log('Dates are available.', response);
        this.showDivSuccessAvailability = true;
         setTimeout(() => {
        this.showDivSuccessAvailability = false;
      }, 5000);

      },
      error: (error) => {
           this.showDiv = true;
           this.errorMessage = error.error.error;
           console.log(error);
            setTimeout(() => {
        this.showDiv = false;
      }, 5000);
           
      }
    }
  );




  } 
}


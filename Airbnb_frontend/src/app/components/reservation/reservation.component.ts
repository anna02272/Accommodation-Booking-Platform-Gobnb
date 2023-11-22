import { Component } from '@angular/core';
import { NgForm } from '@angular/forms';
import { ReservationService } from 'src/app/services/reservation.service';
import { Reservation } from 'src/app/models/reservation';
import { DatePipe } from '@angular/common';


@Component({
  selector: 'app-reservation',
  templateUrl: './reservation.component.html',
  styleUrls: ['./reservation.component.css']
})
export class ReservationComponent {


    accommodation_id!: "03c84b07-8848-11ee-8ba3-0242ac130006";
    //for now this is hardocded since there are no accommodations on frontend
    //accommodation_name!: string;
    //accommodation_location!: string;
    check_in_date?: string;
    check_out_date?: string;
    check_in_time?: number;


    constructor(private reservationService: ReservationService, private datePipe: DatePipe) {
    }


  convertToISOFormat(dateObject?: string): any {
    // Use DatePipe to format the selected date to ISO format
     if (this.datePipe.transform(dateObject, 'yyyy-MM-ddTHH:mm:ssZ') === null){
      return ""
     }
     return this.datePipe.transform(dateObject, 'yyyy-MM-ddTHH:mm:ssZ')
  }
  createReservation(): void {

    if (this.check_in_time === undefined){
      return 
    } 
    else {
      if (this.check_in_time > 24 || this.check_in_time < 1){
        return
      }
    }
    const reservationCreate: Reservation = {
      accommodation_id: "562a302d-8845-11ee-8386-0242ac150007",
      check_in_date: this.check_in_date+`T${this.check_in_time}:00:00Z`,
      check_out_date: this.check_out_date+"T15:00:00Z"
    };

  this.reservationService.createReservation(reservationCreate).subscribe(
    {
      next: (response) => {
        console.log('Reservation created successfully', response);
      },
      error: (error) => {
        console.error('Error creating reservation', error);
      }
    }
  );
    
  }
}
// function ViewChild(arg0: string, arg1:
//    { static: boolean; }): (target: ReservationComponent, propertyKey: "reservationForm") => void {
//   throw new Error('Function not implemented.');
// }


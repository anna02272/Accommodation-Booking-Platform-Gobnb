import { Component, Input, OnInit } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { ReservationService } from 'src/app/services/reservation.service';
import { Reservation } from 'src/app/models/reservation';
import { DatePipe } from '@angular/common';
import { HttpClient } from '@angular/common/http';

@Component({
  selector: 'app-reservation',
  templateUrl: './reservation.component.html',
  styleUrls: ['./reservation.component.css']
})
export class ReservationComponent implements OnInit {
  @Input() accommodationId!: string;
  form!: FormGroup;
  showDiv: boolean = false;
  showDivSuccess: boolean = false;
  showDivSuccessAvailability: boolean = false;
  check_in_date?: string;
  check_out_date?: string;
  check_in_time?: number;
  number_of_guests?: number;
  errorCheck: boolean = false;
  errorCheckGuests: boolean = false;
  errorMessage?: "";
  successMessage?: "Reserved successfully!"
  errorMessage2?: "Please enter your check-in time!"

  constructor(private fb: FormBuilder, private reservationService: ReservationService, private datePipe: DatePipe, private httpClient: HttpClient) {
    this.showDivSuccess = false;
  }

  ngOnInit(): void {
    this.form = this.fb.group({
      check_in_time: [''],
      check_out_date: [''],
      check_in_date: [''],
      number_of_guests: ['']
    });
  }

  convertToISOFormat(dateObject?: string, isCheckOut?: boolean): string {
    const isoFormat = this.datePipe.transform(dateObject, 'yyyy-MM-ddTHH:mm:ss') + 'Z';

    // If it's a check-out date and not provided by the user, assume 15:00:00Z
    if (isCheckOut && !dateObject) {
      return isoFormat.replace(/00:00:00/, '15:00:00');
    }

    return isoFormat;
  }

  createReservation(): void {
    if (this.check_in_time === undefined) {
      this.errorCheck = true;
      return;
    } else {
      if (this.check_in_time > 24 || this.check_in_time < 1) {
        this.errorCheck = true;
        return;
      }
    }

    if (this.number_of_guests === undefined) {
      this.errorCheckGuests = true;
      return;
    } else {
      if (this.number_of_guests < 1) {
        this.errorCheckGuests = true;
        return;
      }
    }

    this.errorCheckGuests = false;
    this.errorCheck = false;

    const reservationCreate: Reservation = {
      accommodation_id: this.accommodationId,
      check_in_date: this.convertToISOFormat(this.check_in_date),
      check_out_date: this.convertToISOFormat(this.check_out_date, true), // Specify it's a check-out date
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
          //this.isHostFeatured();
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
    this.errorCheck = false;

    const checkAvailabilityData = {
      check_in_date: this.convertToISOFormat(this.check_in_date),
      check_out_date: this.convertToISOFormat(this.check_out_date, true), // Specify it's a check-out date
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

  // isHostFeatured() {
  //   //get all ratings for the host getAll returns one json {"ratings": "", "averageRating": ""}
  //   var featured = false;
  //   var averageRating = 0;
  //   this.ratingService.getAll().subscribe(
  //     (response: any) => {
  //       averageRating = response.averageRating;
  //     },
  //     error => {
  //       console.error('Error fetching ratings', error);
  //     }
  //   );
  //   if (averageRating >= 4.7) {
  //     featured = true;
  //   }

  //   var cancelRate = 0;
  //   //get cancelled rate from https://localhost:8000/api/reservations/cancelled/{hostId}
  //   this.httpClient.get('https://localhost:8000/api/reservations/cancelled/' + this.hostId).subscribe(
  //     (response: any) => {
  //       //cancelRate = response to float
  //       cancelRate = response;
  //     },
  //     error => {
  //       console.error('Error fetching cancel rate', error);
  //     }
  //   );
  //   if (cancelRate < 5.0) {
  //     featured = true;
  //   }

  //   var total = 0;
  //   this.httpClient.get('https://localhost:8000/api/reservations/total/' + this.hostId).subscribe(
  //     (response: any) => {
  //       //total = response to float
  //       total = response;
  //     },
  //     error => {
  //       console.error('Error fetching total', error);
  //     }
  //   );
  //   if (total >= 5) {
  //     featured = true;
  //   }

  //   var duration = 0;
  //   this.httpClient.get('https://localhost:8000/api/reservations/duration/' + this.hostId).subscribe(
  //     (response: any) => {
  //       //duration = response to float
  //       duration = response;
  //     },
  //     error => {
  //       console.error('Error fetching duration', error);
  //     }
  //   );
  //   if (duration > 50) {
  //     featured = true;
  //   }

  //   var responseFeatured = false;
  //   this.httpClient.get('https://localhost:8000/api/profile/isFeatured/' + this.hostId).subscribe(
  //       (response: any) => {
  //         responseFeatured = response;
  //       },
  //       error => {
  //         console.error('Error fetching isFeatured', error);
  //       }
  //     );
  //   if (featured) {
  //     if (!responseFeatured) {
  //       //post to https://localhost:8000/api/hosts/featured/{hostId}
  //       this.httpClient.post('https://localhost:8000/api/profile/setFeatured/' + this.hostId, null).subscribe(
  //         (response: any) => {
  //           console.log(response);
  //         },
  //         error => {
  //           console.error('Error featuring host', error);
  //         }
  //       );
  //     }
  //   } else{
  //     if (responseFeatured) {
  //       this.httpClient.post('https://localhost:8000/api/profile/removeFeatured/' + this.hostId, null).subscribe(
  //         (response: any) => {
  //           console.log(response);
  //         },
  //         error => {
  //           console.error('Error removing feature from host', error);
  //         }
  //       );
  //     }
  //   }

  // }
  
}


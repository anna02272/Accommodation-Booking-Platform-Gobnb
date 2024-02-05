import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { GetReservation } from 'src/app/models/GetReservation';
import { RefreshService } from 'src/app/services/refresh.service';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-reservations',
  templateUrl: './reservations.component.html',
  styleUrls: ['./reservations.component.css']
})
export class ReservationsComponent implements OnInit{
reservations: GetReservation[] = [];
notification = { msgType: '', msgBody: '' };
reservationServiceAvailable: boolean = false;

constructor(
  private resService: ReservationService,
  private refreshService: RefreshService,
  private httpClient: HttpClient,
) {}
ngOnInit() {
  this.load();
  this.subscribeToRefresh();
}
load() {
  this.resService.getAll().subscribe((data: GetReservation[]) => {
    this.reservations = data;
    console.log(this.reservations)
    console.log(this.reservations[0])

    console.log(this.reservations[0].accommodation_name)
},
 error => {
if (error.statusText === 'Unknown Error') {
       console.log("here")
       console.log(error)
      this.reservationServiceAvailable = true;
      }
 }
);
}

private subscribeToRefresh() {
  this.refreshService.getRefreshObservable().subscribe(() => {
    this.load();
  });
}
cancelReservation(id: string, accid: string): void {
  this.resService.cancelReservation(id).subscribe(
    () => {
        //this.isHostFeatured(accid);
        this.refreshService.refresh();
        this.notification = { msgType: 'success', msgBody: `Reservation canceled successfully.` };
    },
  error => {
    if (error.status === 400 && error.error && error.error.error) {
      const errorMessage = error.error.error;
      this.notification = { msgType: 'error', msgBody: errorMessage };
    } else {
      this.notification = { msgType: 'error', msgBody: 'An error occurred while processing your request.' };
    }
  }
);
}

// isHostFeatured(accid: string) {
//   var featured = false;
//   var hostId = "";
//     this.httpClient.get('https://localhost:8000/api/accommodations/get/hostid/' + accid).subscribe(
//       (response: any) => {
//         hostId = response.hostId;
//       },
//       error => {
//         console.error('Error fetching hostId', error);
//       }
//     );
//   // var averageRating = 0;
//   // this.ratingService.getAll().subscribe(
//   //   (response: any) => {
//   //     averageRating = response.averageRating;
//   //   },
//   //   error => {
//   //     console.error('Error fetching ratings', error);
//   //   }
//   // );
//   // if (averageRating >= 4.7) {
//   //   featured = true;
//   // }

//   var cancelRate = 0;
//     this.httpClient.get('https://localhost:8000/api/reservations/cancelled/' + hostId).subscribe(
//       (response: any) => {
//         cancelRate = response;
//         alert("cancel rate " + cancelRate);
//       },
//       error => {
//         console.error('Error fetching cancel rate', error);
//       }
//     );
//     if (cancelRate < 5.0) {
//       featured = true;
//     }

//     var total = 0;
//     this.httpClient.get('https://localhost:8000/api/reservations/total/' + hostId).subscribe(
//       (response: any) => {
//         total = response;
//         alert("total " + total);
//       },
//       error => {
//         console.error('Error fetching total', error);
//       }
//     );
//     if (total >= 5) {
//       featured = true;
//     }

//     var duration = 0;
//     this.httpClient.get('https://localhost:8000/api/reservations/duration/' + hostId).subscribe(
//       (response: any) => {
//         duration = response;
//         alert("duration " + duration);
//       },
//       error => {
//         console.error('Error fetching duration', error);
//       }
//     );
//     if (duration > 50) {
//       featured = true;
//     }

//     var responseFeatured = false;
//     this.httpClient.get('https://localhost:8000/api/profile/isFeatured/' + hostId).subscribe(
//         (response: any) => {
//           responseFeatured = response;
//           alert("responseFeatured " + responseFeatured);
//         },
//         error => {
//           console.error('Error fetching isFeatured', error);
//         }
//       );
//     if (featured) {
//       if (!responseFeatured) {
//         //post to https://localhost:8000/api/hosts/featured/{hostId}
//         this.httpClient.post('https://localhost:8000/api/profile/setFeatured/' + hostId, null).subscribe(
//           (response: any) => {
//             console.log(response);
//           },
//           error => {
//             console.error('Error featuring host', error);
//           }
//         );
//       }
//     } else{
//       if (responseFeatured) {
//         this.httpClient.post('https://localhost:8000/api/profile/removeFeatured/' + hostId, null).subscribe(
//           (response: any) => {
//             console.log(response);
//           },
//           error => {
//             console.error('Error removing feature from host', error);
//           }
//         );
//       }
//     }

// }
}

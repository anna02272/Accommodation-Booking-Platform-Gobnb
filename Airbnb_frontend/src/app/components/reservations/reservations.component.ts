import { HttpClient } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { concatMap } from 'rxjs/operators';
import { GetReservation } from 'src/app/models/GetReservation';
import { RatingService } from 'src/app/services/rating.service';
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
  accommodationId!: string;
  hostId!: string;
  hostEmail!: string;
  hostIdP!: string;
  hostFeatured!: boolean;

constructor(
  private resService: ReservationService,
  private refreshService: RefreshService,
  private httpClient: HttpClient,
  private ratingService: RatingService
) {}
ngOnInit() {
  this.load();
  this.subscribeToRefresh();

  

}

featuredCheckData(){
  this.httpClient.get('https://localhost:8000/api/accommodations/get/hostid/' + this.accommodationId).pipe(
      concatMap((response: any) => {
        this.hostId = response.hostId;
        //alert("accid " + this.accommodationId + " hostid " + this.hostId);
        return this.httpClient.get('https://localhost:8000/api/users/getById/' + this.hostId);
      }),
      concatMap((response: any) => {
        this.hostEmail = response.user.email;
        //alert("hostEmail " + this.hostEmail);
        return this.httpClient.get('https://localhost:8000/api/profile/getUser/' + this.hostEmail);
      }),
      concatMap((response: any) => {
        this.hostIdP = response.user.id;
        //alert("hostIdP " + this.hostIdP);
        return this.httpClient.get('https://localhost:8000/api/profile/isFeatured/' + this.hostIdP);
      })
    ).subscribe(
      (response: any) => {
        this.hostFeatured = response.featured;
        //alert("hostFeatured " + this.hostFeatured);
      },
      error => {
        console.error('Error', error);
      }
    );
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
  
  //this.isHostFeatured();
  this.accommodationId = accid;
  this.featuredCheckData();
  //this.isHostFeatured();
  //first run featuredCheckData(), once it finishes completely run isHostFeatured()
  setTimeout(() => {
    this.isHostFeatured();
  }, 1500);

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

isHostFeatured() {
  //alert("isHostFeatured");
  var featured = false;
  
  var averageRating = 0;
  this.ratingService.getAll().subscribe(
    (response: any) => {
      averageRating = response.averageRating;
    },
    error => {
      console.error('Error fetching ratings', error);
    }
  );
  // if (averageRating >= 4.7) {
  //   featured = true;
  // } else{
  //   featured = false;
  // }

  var cancelRate = 0;
  this.httpClient.get('https://localhost:8000/api/reservations/cancelled/' + this.hostId).subscribe(
    (response: any) => {
      cancelRate = response;
      // if (cancelRate < 5.0) {
      //   featured = true;
      //   alert("cancel rate " + cancelRate)
      // } else{
      //   featured = false;
      //   alert("FALSE cancel rate " + cancelRate)
      // }
      //alert("cancel rate " + cancelRate);
    },
    error => {
      console.error('Error fetching cancel rate', error);
    }
  );
  

  var total = 0;
  this.httpClient.get('https://localhost:8000/api/reservations/total/' + this.hostId).subscribe(
    (response: any) => {
      total = response;
      // if (total >= 5) {
      //   featured = true;
      //   alert("total " + total)
      // } else{
      //   featured = false;
      //   alert("FALSE total " + total)
      // }
      //alert("total " + total);
    },
    error => {
      console.error('Error fetching total', error);
    }
  );
  

  var duration = 0;
  this.httpClient.get('https://localhost:8000/api/reservations/duration/' + this.hostId).subscribe(
    (response: any) => {
      duration = response;
      // if (duration > 50) {
      //   featured = true;
      //   alert("duration " + duration)
      // } else{
      //   featured = false;
      //   alert("FALSE duration " + duration)
      // }
      //alert("duration " + duration);
    },
    error => {
      console.error('Error fetching duration', error);
    }
  );
  

  
  //var responseFeatured = false;
  
  setTimeout(() => {
    
    if((averageRating >= 4.5) && (cancelRate < 5.0) && (total >= 5) && (duration > 50)){
      featured = true;
    } else{
      featured = false;
    }

    //alert("featured " + featured);
    if (this.hostFeatured) {
      if (!featured) {
        //post to https://localhost:8000/api/hosts/featured/{hostId}
        this.httpClient.post('https://localhost:8000/api/profile/setUnfeatured/' + this.hostIdP, null).subscribe(
          (response: any) => {
            console.log(response);
            console.log("Host is now not featured")
            //alert("response set unfeatured " + response);
          },
          error => {
            console.error('Error featuring host', error);
            //alert("error set featured " + error);
          }
        );
      }
    } else{
      if (featured) {
        this.httpClient.post('https://localhost:8000/api/profile/setFeatured/' + this.hostIdP, null).subscribe(
          (response: any) => {
            console.log(response);
            console.log("Host is now featured")
            //alert("response set featured " + response);
          },
          error => {
            console.error('Error removing feature from host', error);
            //alert("error set unfeatured " + error);
          }
        );
      }
    }

  }, 1000);

}

}

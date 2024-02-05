import { HttpClient } from '@angular/common/http';
import { Component, AfterViewInit, Input, SimpleChanges } from '@angular/core';
import { concatMap } from 'rxjs/operators';
import { RatingItem } from 'src/app/models/rateHost';
import { UserService } from 'src/app/services';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-host',
  templateUrl: './rate-host.component.html',
  styleUrls: ['./rate-host.component.css']
})
export class RateHostComponent implements AfterViewInit {
  @Input() hostId!: string;
  hostId2: string = "65bccbd379154f7558b148f7";
  notification1 = { msgType: '', msgBody: '' };
  selectedRating: number | null = null;
  ratings: RatingItem[] = [];
  ratingServiceAvailable: boolean = false;
  accommodationId!: string;
  hostEmail!: string;
  hostIdP!: string;
  hostFeatured!: boolean;

  constructor(
    private ratingService: RatingService,
    private userService: UserService,
    private httpClient: HttpClient
  ) {}

  ngOnInit(): void {
    
  }


  checkFeaturedData(){

    this.httpClient.get('https://localhost:8000/api/users/getById/' + this.hostId).pipe(
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
          //alert("hostFeatured " + response.featured + this.hostFeatured);
        },
        error => {
          console.error('Error', error);
        }
      );

  }

  ngOnChanges(changes: SimpleChanges): void {
    if ('hostId' in changes) {
      this.fetchRating();
    }
  }

  fetchRating(): void {
    if (!this.hostId) {
      return;
    }

    this.ratingService.getByHostAndGuest(this.hostId).subscribe(
      (response: any) => {
        if (response.ratings && response.ratings.length > 0) {
          this.selectedRating = response.ratings[0].rating;
          this.updateStars();
        }
      },
      error => {
        console.error('Error fetching rating', error);
      }
    );
  }


  ngAfterViewInit() {
    const resetStarsButton = document.getElementById('resetStars');
    if (resetStarsButton) {
      resetStarsButton.addEventListener('click', () => {
        this.resetStars();
      });
    }

    const stars = document.getElementsByName('hostRating') as NodeListOf<HTMLInputElement>;
    stars.forEach((star: HTMLInputElement) => {
      star.addEventListener('click', () => {
        this.selectedRating = Number(star.value);
        this.rateHost();
      });
    });

  }

  getUserId() {
    return this.userService.currentUser?.user.ID;
  }

  resetStars(): void {
    const stars = document.getElementsByName('hostRating') as NodeListOf<HTMLInputElement>;
    stars.forEach((star: HTMLInputElement) => {
      star.checked = false;
    });
    this.selectedRating = null;
  }

  rateHost(): void {
    setTimeout(() => {
      this.checkFeaturedData();
    }, 1000);
    setTimeout(() => {
      this.isHostFeatured();
    }, 3500);
    if (!this.hostId || this.selectedRating === null) {
      console.error('Host ID or rating is not provided.');
      return;
    }

    this.ratingService.rateHost(this.hostId, this.selectedRating).subscribe(
      response => {
        this.notification1 = { msgType: 'success', msgBody: 'Rating successfully submitted' };
        
        //this.isHostFeatured();

      },
    error => {
      //this.isHostFeatured();
      if (error.status === 400 && error.error && error.error.error) {
        const errorMessage = error.error.error;
        this.notification1 = { msgType: 'error', msgBody: errorMessage };
      }
      else if (error.statusText === 'Unknown Error') {
       console.log("here")
       console.log(error)
      this.ratingServiceAvailable = true;
      }
      
      else {
        this.notification1 = { msgType: 'error', msgBody: 'An error occurred while processing your request.' };
      }
      
    }
  );
  }

  updateStars(): void {
    const stars = document.getElementsByName('hostRating') as NodeListOf<HTMLInputElement>;
    stars.forEach((star: HTMLInputElement) => {
      star.checked = Number(star.value) === this.selectedRating;
    });
  }

  deleteRating(): void {
    if (!this.hostId) {
      console.error('Host ID is not provided.');
      return;
    }
    this.ratingService.deleteRating(this.hostId).subscribe(
      response => {
        this.notification1 = { msgType: 'success', msgBody: 'Rating successfully deleted' };
      },
      error => {
        if (error.status === 400 && error.error && error.error.error) {
          const errorMessage = error.error.error;
          this.notification1 = { msgType: 'error', msgBody: errorMessage };
        } else {
          this.notification1 = { msgType: 'error', msgBody: 'An error occurred while processing your request.' };
        }
      }
    );
  }

  isHostFeatured() {
    //alert("hostid " + this.hostId + " hostidp " + this.hostIdP);
    //alert("isHostFeatured");
    var featured = false;
    
    // var averageRating = 0;
    // this.ratingService.getAll().subscribe(
    //   (response: any) => {
    //     averageRating = response.averageRating;
    //   },
    //   error => {
    //     console.error('Error fetching ratings', error);
    //   }
    // );
    // if (averageRating >= 4.7) {
    //   featured = true;
    // }

    var cancelRate = 0;
    this.httpClient.get('https://localhost:8000/api/reservations/cancelled/' + this.hostId).subscribe(
      (response: any) => {
        cancelRate = response;
        //alert("cancel rate " + cancelRate);
      },
      error => {
        console.error('Error fetching cancel rate', error);
      }
    );
    if (cancelRate < 5.0) {
      featured = true;
    }

    var total = 0;
    this.httpClient.get('https://localhost:8000/api/reservations/total/' + this.hostId).subscribe(
      (response: any) => {
        total = response;
        //alert("total " + total);
      },
      error => {
        console.error('Error fetching total', error);
      }
    );
    if (total >= 5) {
      featured = true;
    }

    var duration = 0;
    this.httpClient.get('https://localhost:8000/api/reservations/duration/' + this.hostId).subscribe(
      (response: any) => {
        duration = response;
        //alert("duration " + duration);
      },
      error => {
        console.error('Error fetching duration', error);
      }
    );
    if (duration > 50) {
      featured = true;
    }

    //alert("isFeatured " + this.hostFeatured + " nowFeatured " + featured);

    //alert("featured " + featured);
    //var responseFeatured = false;
    
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

  }

}
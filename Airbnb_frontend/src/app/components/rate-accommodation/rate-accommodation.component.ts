import { Component, AfterViewInit, Input, SimpleChanges } from '@angular/core';
import { RatingItem } from 'src/app/models/rateAccommodation';
import { UserService } from 'src/app/services';
import { RatingService } from 'src/app/services/rating.service';

@Component({
  selector: 'app-rate-accommodation',
  templateUrl: './rate-accommodation.component.html',
  styleUrls: ['./rate-accommodation.component.css']
})
export class RateAccommodationComponent implements AfterViewInit {
  @Input() accommodationId!: string;
  notification2 = { msgType: '', msgBody: '' };
  selectedRating: number | null = null;
  ratings: RatingItem[] = [];

  constructor(
    private ratingService: RatingService,
    private userService: UserService,
  ) {}

  ngOnChanges(changes: SimpleChanges): void {
    if ('accommodationId' in changes) {
      this.fetchRating();
    }
  }

  fetchRating(): void {
    if (!this.accommodationId) {
      return;
    }

    this.ratingService.getByAccommodationAndGuest(this.accommodationId).subscribe(
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
    const resetStarsButton = document.getElementById('resetStar');
    if (resetStarsButton) {
      resetStarsButton.addEventListener('click', () => {
        this.resetStars();
      });
    }

    const stars = document.getElementsByName('accommodationRating') as NodeListOf<HTMLInputElement>;
    stars.forEach((star: HTMLInputElement) => {
      star.addEventListener('click', () => {
        this.selectedRating = Number(star.value);
        this.rateAccommodation();
      });
    });

  }

  getUserId() {
    return this.userService.currentUser?.user.ID;
  }

  resetStars(): void {
    const stars = document.getElementsByName('accommodationRating') as NodeListOf<HTMLInputElement>;
    stars.forEach((starr: HTMLInputElement) => {
      starr.checked = false;
    });
    this.selectedRating = null;
  }

  rateAccommodation(): void {
    if (!this.accommodationId || this.selectedRating === null) {
      console.error('Accommodation ID or rating is not provided.');
      return;
    }

    this.ratingService.rateAccommodation(this.accommodationId, this.selectedRating).subscribe(
      response => {
        this.notification2 = { msgType: 'success', msgBody: 'Rating accommodation successfully submitted' };
      },
    error => {
      if (error.status === 400 && error.error && error.error.error) {
        const errorMessage = error.error.error;
        this.notification2 = { msgType: 'error', msgBody: errorMessage };
      } else {
        this.notification2 = { msgType: 'error', msgBody: 'An error occurred while processing your request.' };
      }
    }
  );
  }

  updateStars(): void {
    const stars = document.getElementsByName('accommodationRating') as NodeListOf<HTMLInputElement>;
    stars.forEach((star: HTMLInputElement) => {
      star.checked = Number(star.value) === this.selectedRating;
    });
  }

  deleteRatingAccommodation(): void {
    console.log(this.accommodationId)
    if (!this.accommodationId) {
      console.error('Accommodation ID is not provided.');
      return;
    }
    this.ratingService.deleteRatingAccommodation(this.accommodationId).subscribe(
      response => {
        this.notification2 = { msgType: 'success', msgBody: 'Rating successfully deleted' };
      },
      error => {
        if (error.status === 400 && error.error && error.error.error) {
          const errorMessage = error.error.error;
          this.notification2 = { msgType: 'error', msgBody: errorMessage };
        } else {
          this.notification2 = { msgType: 'error', msgBody: 'An error occurred while processing your request.' };
        }
      }
    );
  }

}
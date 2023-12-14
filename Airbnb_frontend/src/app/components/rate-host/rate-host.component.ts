import { Component, AfterViewInit, Input, SimpleChanges } from '@angular/core';
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
  notification = { msgType: '', msgBody: '' };
  selectedRating: number | null = null;
  ratings: RatingItem[] = [];

  constructor(
    private ratingService: RatingService,
    private userService: UserService
  ) {}

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
    if (!this.hostId || this.selectedRating === null) {
      console.error('Host ID or rating is not provided.');
      return;
    }

    this.ratingService.rateHost(this.hostId, this.selectedRating).subscribe(
      response => {
        this.notification = { msgType: 'success', msgBody: 'Rating successfully submitted' };
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
        this.notification = { msgType: 'success', msgBody: 'Rating successfully deleted' };
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

}
import { Component, AfterViewInit, Input } from '@angular/core';
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

  constructor(
    private ratingService: RatingService
  ) {}

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
        this.notification = { msgType: 'error', msgBody: 'Error submitting rating' };
      }
    );
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
        this.notification = { msgType: 'error', msgBody: 'Error deleting rating' };
      }
    );
  }
}

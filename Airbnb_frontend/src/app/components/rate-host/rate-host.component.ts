import { Component, AfterViewInit } from '@angular/core';

@Component({
  selector: 'app-rate-host',
  templateUrl: './rate-host.component.html',
  styleUrls: ['./rate-host.component.css']
})
export class RateHostComponent implements AfterViewInit {

  ngAfterViewInit() {
    const resetStarsButton = document.getElementById('resetStars');
    if (resetStarsButton) {
      resetStarsButton.addEventListener('click', () => {
        const stars = document.getElementsByName('hostRating') as NodeListOf<HTMLInputElement>;
        stars.forEach((star: HTMLInputElement) => {
          star.checked = false;
        });
      });
    }
  }

}

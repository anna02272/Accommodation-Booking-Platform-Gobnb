import { Component } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.css']
})
export class SearchComponent {

  constructor(
    private router: Router
    ) {}

  search() {
    const location = (document.getElementById('location') as HTMLInputElement).value;
    const guests = (document.getElementById('guests') as HTMLInputElement).value;
    const startDate = (document.getElementById('startDate') as HTMLInputElement).value;
    const endDate = (document.getElementById('endDate') as HTMLInputElement).value; 
    const tv = (document.getElementById('tv') as HTMLInputElement).checked;
    const wifi = (document.getElementById('wifi') as HTMLInputElement).checked;
    const ac = (document.getElementById('ac') as HTMLInputElement).checked;
    // const amenities = {
    //   'TV': tv,
    //   'WiFi': wifi,
    //   'AC': ac
    // };
    
    const minPrice = parseInt((document.getElementById('min_price') as HTMLInputElement).value, 10);
    const maxPrice = parseInt((document.getElementById('max_price') as HTMLInputElement).value, 10);
    console.log(minPrice);
    console.log(maxPrice);

    if ((!Number.isNaN(minPrice) && Number.isNaN(maxPrice)) || (Number.isNaN(minPrice) && !Number.isNaN(maxPrice))) {
      alert('Please enter both prices!');
      return;
    }

    if (minPrice > maxPrice) {
      alert('Min price must be less than max price');
      return;
    }

    if(!Number.isNaN(minPrice) && !Number.isNaN(maxPrice)){
      if(startDate == '' && endDate == ''){
        alert('Please enter dates!');
        return;
      }
    }

    if ((startDate != '' && endDate == '') || (startDate == '' && endDate != '')) {
      alert('Please enter both dates!');
      return;
    }
    //if startDate is after endDate
    if ((startDate != '' && endDate != '') && (startDate > endDate)) {
      alert('Start date must be before end date');
      return;
    }
    window.location.href = '/home?location=' + location + '&guests=' + guests + '&start_date=' + startDate + '&end_date=' + endDate + '&tv=' + tv + '&wifi=' + wifi + '&ac=' + ac + '&min_price=' + minPrice + '&max_price=' + maxPrice;
  }

  isButtonVisible(){
    if (window.location.search){
      return true;
    }
    return false;
  }

  goHome(){
    window.location.href = '/home';
  }

  advanced(){
    //show or hide div advanced_search
    var x = document.getElementById("advanced_search");
    if(x){
    if (x.style.display === "none") {
      x.style.display = "block";
      //change button text
      var advancedBtn = document.getElementById("advanced-btn");
      if (advancedBtn !== null) {
        advancedBtn.innerHTML = "Advanced UP";
      }
    } else {
      x.style.display = "none";
      //change button text
      var advancedBtn = document.getElementById("advanced-btn");
      if (advancedBtn !== null) {
        advancedBtn.innerHTML = "Advanced DOWN";
      }
    }
  }
  }

}

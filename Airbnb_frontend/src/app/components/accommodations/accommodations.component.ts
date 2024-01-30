import { ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { Accommodation } from 'src/app/models/accommodation';
import { Recommendation } from 'src/app/models/recommendation';
import { UserService } from 'src/app/services';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { RefreshService } from 'src/app/services/refresh.service';
import { DomSanitizer } from '@angular/platform-browser';
import { forkJoin } from 'rxjs/internal/observable/forkJoin';
import { ReservationService } from 'src/app/services/reservation.service';

@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit {
  accommodations: Accommodation[] = [];
  accommodation!: Accommodation;
  recommendations: Recommendation[]=[];
  recommendation! : Recommendation;
  accServiceAvailable: boolean = false;
  showErrorDiv: boolean = false;
  showSuccessMsg: boolean = false;
  errorMsg?: string;
  images!: any[];
  coverImage: string = ''; 
  currentIndex: number = 0;
  gtag?: Function;
  id:any;


  constructor(
    private accService: AccommodationService,
    private refreshService: RefreshService,
    private userService: UserService,
    private sanitizer: DomSanitizer,
    private reservationService: ReservationService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    this.userService.getMyInfo().subscribe(
      () => {
        this.loadGuestId();
        this.loadByHost();
        this.subscribeToRefresh();
      },
      () => {
        this.load();
        // this.loadRecommendation();
        this.subscribeToRefresh();
      },
      () => {
        this.loadAccommodationImages();
        this.subscribeToRefresh();
      }
      
    );
  }
  
  loadByHost() {
    const userRole = this.userService.currentUser?.user.userRole;

    if (userRole === 'Host') {
      this.accService.getByHost(this.userService.currentUser.user.id).subscribe((data: Accommodation[]) => {
        this.accommodations = data;
        this.loadAccommodationImages();
      });
    } else {
      this.load();
      this.loadAccommodationImages();
      // this.loadRemmendationAccommodation();

    }
  }
  loadGuestId(){
    this.id=this.userService.currentUser?.user.id
  }
  loadRecommendation(){

    this.accService.getAllRecommendation(this.id).subscribe((dataa:Recommendation[])=>{
      this.recommendations=dataa;
      console.log(dataa)
      if(this.recommendations.length){
        var notif = document.getElementById("notif");
        notif!.style.display = "none";
      } else{
        var notif = document.getElementById("notif");
        notif!.style.display = "block";
      }
    })
  }
  load() {
    if (window.location.search){
      //alert("here")
      //get the params from the url
      var urlParams = new URLSearchParams(window.location.search);
      //get the values from the params
      var location = urlParams.get('location');
      var guests = urlParams.get('guests');
      var start_date = urlParams.get('start_date');
      var end_date = urlParams.get('end_date');
      var tv = urlParams.get('tv');
      var wifi = urlParams.get('wifi');
      var ac = urlParams.get('ac');
      var min_price = urlParams.get('min_price');
      var max_price = urlParams.get('max_price');
      //var price_type = urlParams.get('price_type');
      this.accService.getSearch(location, guests, start_date, end_date, tv, wifi, ac, min_price, max_price).subscribe((data: Accommodation[]) => {
      var arr = data;
      //var arr2: Accommodation[] = [];
      this.accommodations = arr;
      if(this.accommodations.length){
        var notif = document.getElementById("notif");
        notif!.style.display = "none"; 
      } else{
        var notif = document.getElementById("notif");
        notif!.style.display = "block";
      }
      this.loadAccommodationImages()
      //this.accommodations = [...this.accommodations];
      this.cdr.detectChanges();
        
      if(start_date != "" && end_date != ""){
        this.accommodations = [];
        for (let acc of arr){
          var check = this.checkAvailability(acc, start_date, end_date, min_price, max_price)
          //, price_type)
        }
      }

      
      
        //this.loadAccommodationImages();

        //this.accommodations = arr2;
      });
    }
    else{
    this.loadRecommendation();

    this.accService.getAll().subscribe((data: Accommodation[]) => {
      this.accommodations = data;
      if(this.accommodations.length){
        var notif = document.getElementById("notif");
        notif!.style.display = "none";
      } else{
        var notif = document.getElementById("notif");
        notif!.style.display = "block";
      }
      this.loadAccommodationImages();
    },
    
    (error) => {
    if (error.statusText === 'Unknown Error') {
       console.log("here")
       console.log(error)
      this.accServiceAvailable = true;
      }
  }
    
    );
    }
  }

  checkAvailability(acc: any, startDate: any, endDate: any, min_price: any, max_price: any): boolean {

    var errorCheck = false;

    const checkAvailabilityData = {
      check_in_date: startDate + "T00:00:00Z",
      check_out_date: endDate + "T00:00:00Z",
    };

    this.reservationService.checkAvailability(checkAvailabilityData, acc._id).subscribe(
      {
        next: (response) => {
          console.log('Dates are available.', response);
          //alert(acc._Id + " is available")
          if(min_price != "NaN" && max_price != "NaN"){
            var check = this.checkPrice(acc, startDate, endDate, min_price, max_price)
            return
          }
          var notif = document.getElementById("notif");
          //notif!.style.display = "none"; 
          this.accommodations.push(acc);
          this.loadAccommodationImages();
          if(this.accommodations.length){
            var notif = document.getElementById("notif");
            notif!.style.display = "none"; 
          } else{
            var notif = document.getElementById("notif");
            notif!.style.display = "block";
          }
          this.cdr.detectChanges();
          //this.showDivSuccessAvailability = true;
          errorCheck = true;
        //    setTimeout(() => {
        //   //this.showDivSuccessAvailability = false;
        //   errorCheck = false;
        // }, 5000);

        },
        error: (error) => {
            //this.showDiv = true;
            //this.errorMessage = error.error.error;
            console.log(error);
            if(this.accommodations.length){
              var notif = document.getElementById("notif");
              notif!.style.display = "none"; 
            } else{
              var notif = document.getElementById("notif");
              notif!.style.display = "block";
            }
            //alert(acc._Id + " is not available")
        //       setTimeout(() => {
        //   //this.showDiv = false;
        //   errorCheck = false;
        // }, 5000);
            
        }
      });

    return errorCheck;

  }

  checkPrice(acc: any, start_date: any, end_date: any, min_price: any, max_price: any): boolean {

    //alert("hereTRIED")

    var errorCheck = false;

    const checkPriceData = {
      check_in_date: start_date + "T00:00:00Z",
      check_out_date: end_date + "T00:00:00Z",
    };

    this.reservationService.checkPrice(checkPriceData, acc._id).subscribe(
      {
        next: (response) => {
          //response is an array of maps price:pricetype, i need an array with only the prices
          var prices = [];
          for (let res of response){
            prices.push(res.price)
          }
          var min = Math.min(...prices)
          //alert("min_price: " + min_price)
          //alert("min: " + min)
          var max = Math.max(...prices)
          //alert("max_price: " + max_price)
          //alert("max: " + max)
          if(min_price <= min && max_price >= max){
            //alert(acc._Id + " is available")
            var notif = document.getElementById("notif");
            //notif!.style.display = "none"; 
            this.accommodations.push(acc);
            this.loadAccommodationImages();
            if(this.accommodations.length){
              var notif = document.getElementById("notif");
              notif!.style.display = "none"; 
            } else{
              var notif = document.getElementById("notif");
              notif!.style.display = "block";
            }
            this.cdr.detectChanges();
            //this.showDivSuccessAvailability = true;
            errorCheck = true;
          }



        //   console.log('Dates are available.', response);
        //   //alert(acc._Id + " is available")
        //   this.accommodations.push(acc);
        //   this.loadAccommodationImages();
        //   this.cdr.detectChanges();
        //   //this.showDivSuccessAvailability = true;
        //   errorCheck = true;
        // //    setTimeout(() => {
        // //   //this.showDivSuccessAvailability = false;
        // //   errorCheck = false;
        // // }, 5000);

        },
        error: (error) => {
            //this.showDiv = true;
            //this.errorMessage = error.error.error;
            console.log(error);
            //alert(acc._Id + " is not available")
        //       setTimeout(() => {
        //   //this.showDiv = false;
            errorCheck = false;
            if(this.accommodations.length){
              var notif = document.getElementById("notif");
              notif!.style.display = "none"; 
            } else{
              var notif = document.getElementById("notif");
              notif!.style.display = "block";
            }
        // }, 5000);
            
        }
      });

    return errorCheck;

  }

 loadAccommodationImages() {
  const imageRequests = this.accommodations.map(accommodation =>
    this.getImages(accommodation._id, accommodation)
  );

  forkJoin(imageRequests).subscribe(() => {
  });
}

  getRole() {
    return this.userService.currentUser?.user.userRole;
  }

  deleteAccommodation(accId: string) {
    this.accService.deleteAccommodation(accId).subscribe(
      () => {
        const index = this.accommodations.findIndex((acc) => acc._id === accId);
        if (index !== -1) {
          this.accommodations.splice(index, 1);
        }

        this.showSuccessMsg = true;
        this.scrollToTop();

        setTimeout(() => {
          this.showSuccessMsg = false;
        }, 5000);
      },
      (error) => {
        console.error('Error deleting accommodation:', error);
        this.errorMsg = error.error.error;
        this.showErrorDiv = true;

        this.scrollToTop();

        setTimeout(() => {
          this.showErrorDiv = false;
        }, 5000);
      }
    );
  }


 getImages(accId: string, accommodation: Accommodation) {
    this.accService.fetchAccImages(accId).subscribe(
      (images: any[]) => {
        this.images = images.map(image => this.arrayBufferToBase64(image.data));
        for (let im of images) {
          let objectURL = 'data:image/png;base64,' + im.data;
          let imageTest = this.sanitizer.bypassSecurityTrustUrl(objectURL);
          this.images[images.indexOf(im)] = imageTest;
        }

        if (this.images.length > 0) {
          accommodation.coverImage = this.images[0];

        } else {
          accommodation.coverImage = ''; 
        }
      },
      (error) => {
        console.error('Error fetching images:', error);
        accommodation.coverImage = ''; 
      }
    );
  }

  private subscribeToRefresh() {
    this.refreshService.getRefreshObservable().subscribe(() => {
      this.load();
      this.loadAccommodationImages();

    });
  }

  scrollToTop() {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  private arrayBufferToBase64(buffer: ArrayBuffer): string {
    let binary = '';
    const bytes = new Uint8Array(buffer);
    const len = bytes.byteLength;
    for (let i = 0; i < len; i++) {
      binary += String.fromCharCode(bytes[i]);
    }
    return 'data:image/jpeg;base64,' + btoa(binary);
  }


  
}


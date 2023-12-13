import { Component, OnInit } from '@angular/core';
import { Accommodation } from 'src/app/models/accommodation';
import { User } from 'src/app/models/user';
import { UserService } from 'src/app/services';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { RefreshService } from 'src/app/services/refresh.service';
import { map, tap } from 'rxjs/operators';
import { UrlSerializer } from '@angular/router';

@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit{
  accommodations: Accommodation[] = [];
  loggedHostAccommodations: any[] = [];
  loggedHost?: any
  showErrorDiv: boolean = false
  showSuccessMsg: boolean = false
  errorMsg?: string;

  constructor(
    private accService: AccommodationService,
    private refreshService: RefreshService,
    private userService: UserService
  ) {}

  ngOnInit() {
      this.load();
      this.subscribeToRefresh();

    if (localStorage.getItem("jwt") !== ""){
      console.log("here")
    this.userService.getMyInfo().pipe(
        tap(user =>{ 
          this.loggedHost = user
          this.loadHostAcc();
          this.refreshService.getRefreshObservable().subscribe(() => {
            this.loadHostAcc();
            });
          console.log(this.accommodations)

        })
      )
      .subscribe();
    }



  
  }

  load() {
      this.accService.getAll().subscribe((data: Accommodation[]) => {
        this.accommodations = data;
  });
  
}

loadHostAcc(){
    this.accService.getByHost(this.loggedHost.user.id).subscribe((data:any) => {
        this.loggedHostAccommodations = data.accommodations;
        let tempArray = []
        for (let acc of this.loggedHostAccommodations){
            tempArray.push(acc._id)
        }
        this.loggedHostAccommodations = tempArray
        console.log(this.loggedHostAccommodations)

                  
        for (let acc of this.accommodations){
          if (this.loggedHostAccommodations.indexOf(acc._id) != -1){
            var index = this.accommodations.indexOf(acc);
            acc.flagCanDelete = true;
            this.accommodations[index] = acc

          }
        }
  });
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



private subscribeToRefresh() {
  this.refreshService.getRefreshObservable().subscribe(() => {
    this.load();
  });
}

scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

}

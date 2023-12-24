import { Component, OnInit } from '@angular/core';
import { Accommodation } from 'src/app/models/accommodation';
import { UserService } from 'src/app/services';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { RefreshService } from 'src/app/services/refresh.service';
import { DomSanitizer } from '@angular/platform-browser';
import { forkJoin } from 'rxjs/internal/observable/forkJoin';

@Component({
  selector: 'app-accommodations',
  templateUrl: './accommodations.component.html',
  styleUrls: ['./accommodations.component.css']
})
export class AccommodationsComponent implements OnInit {
  accommodations: Accommodation[] = [];
  accommodation!: Accommodation;
  showErrorDiv: boolean = false;
  showSuccessMsg: boolean = false;
  errorMsg?: string;
  images!: any[];
  coverImage: string = ''; 
  currentIndex: number = 0;


  constructor(
    private accService: AccommodationService,
    private refreshService: RefreshService,
    private userService: UserService,
    private sanitizer: DomSanitizer
  ) {}

  ngOnInit() {
    this.userService.getMyInfo().subscribe(
      () => {
        this.loadByHost();
        this.subscribeToRefresh();
      },
      () => {
        this.load();
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

    }
  }

  load() {
    this.accService.getAll().subscribe((data: Accommodation[]) => {
      this.accommodations = data;
      this.loadAccommodationImages();
    });
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


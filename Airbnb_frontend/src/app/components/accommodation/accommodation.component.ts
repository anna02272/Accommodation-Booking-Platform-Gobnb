import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Accommodation } from 'src/app/models/accommodation';
import { UserService } from 'src/app/services';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { DomSanitizer } from '@angular/platform-browser';

@Component({
  selector: 'app-accommodation',
  templateUrl: './accommodation.component.html',
  styleUrls: ['./accommodation.component.css']
})
export class AccommodationComponent implements OnInit {
  accId!: string; 
  hostId!: string;
  accommodation!: Accommodation;
  tv!: boolean;
  wifi!: boolean;
  ac!: boolean;
  am_map!: Map<string, boolean>;
  images!: any[];
  currentImage: string = ''; 
  currentIndex: number = 0;
  
  constructor( 
    private userService: UserService,
    private accService : AccommodationService,
    private route: ActivatedRoute ,
    private sanitizer: DomSanitizer
    ) 
  { }
 

  ngOnInit(): void {
    this.accId = this.route.snapshot.paramMap.get('id')!;
    this.accService.getById(this.accId).subscribe((accommodation: Accommodation) => {
      this.accommodation = accommodation;
      this.hostId = accommodation.host_id;
      this.am_map = new Map<string, boolean>();
      this.am_map = Object.entries(this.accommodation.accommodation_amenities).reduce((map, [key, value]) => map.set(key, value), new Map<string, boolean>());
      this.tv = this.am_map.get('TV')!;
      this.wifi = this.am_map.get('WiFi')!;
      this.ac = this.am_map.get('AC')!;
    });

    this.getImages(this.accId);
  }


getImages(accId: string) {
  this.accService.fetchAccImages(accId).subscribe(
   (images: any[]) => {
      this.images = images.map(image => this.arrayBufferToBase64(image.data));
      for (let im of images){
        console.log(im.data);
         let objectURL = 'data:image/png;base64,' + im.data;
        let imageTest = this.sanitizer.bypassSecurityTrustUrl(objectURL);
        this.images[images.indexOf(im)] = imageTest;
      }
      if (this.images.length > 0) {
        this.currentImage = this.images[0];
      }
    },
    (error) => {
    console.error('Error fetching images:', error);
     
    }
  );
}

arrayBufferToBase64(buffer: ArrayBuffer): string {
  let binary = '';
  const bytes = new Uint8Array(buffer);
  const len = bytes.byteLength;
  for (let i = 0; i < len; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return 'data:image/jpeg;base64,' + btoa(binary);
}

prevImage() {
  if (this.currentIndex > 0) {
    this.currentIndex--;
    this.currentImage = this.images[this.currentIndex];
  }
}

nextImage() {
  if (this.currentIndex < this.images.length - 1) {
    this.currentIndex++;
    this.currentImage = this.images[this.currentIndex];
  }
}

  getRole() {
    return this.userService.currentUser?.user.userRole;
  }
}

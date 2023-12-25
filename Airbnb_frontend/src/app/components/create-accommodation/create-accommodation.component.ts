import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Component } from '@angular/core';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { AccDataService } from 'src/app/services/acc-data-service';
import { ActivatedRoute, Router } from '@angular/router';
import { UserService } from 'src/app/services';

@Component({
  selector: 'app-create-accommodation',
  templateUrl: './create-accommodation.component.html',
  styleUrls: ['./create-accommodation.component.css']
})
export class CreateAccommodationComponent {
  notification = { msgType: '', msgBody: '' };
  images: File[] = [];

  //@ViewChild('fileInput') fileInput: ElementRef; // Access the file input element

  // constructor(
  //   private accService: AccommodationService
  // ) {
  // }
  constructor(private dataService: AccDataService,
    private router: Router,
    private userService: UserService,
    private accService: AccommodationService
    ) {}

    handleImageChange(event: any): void {
    this.images = event.target.files;
  }



  onSubmit() {
    const name = (document.getElementById('name') as HTMLInputElement).value;
    const location = (document.getElementById('location') as HTMLInputElement).value;
    //const amenities = (document.getElementById('amenities') as HTMLTextAreaElement).value;
    const tv = (document.getElementById('tv') as HTMLInputElement).checked;
    const wifi = (document.getElementById('wifi') as HTMLInputElement).checked;
    const ac = (document.getElementById('ac') as HTMLInputElement).checked;
    const minGuests = parseInt((document.getElementById('minGuests') as HTMLInputElement).value, 10);
    const maxGuests = parseInt((document.getElementById('maxGuests') as HTMLInputElement).value, 10);
    const amenities = {
      'TV': tv,
      'WiFi': wifi,
      'AC': ac
    };
    
    //const files: FileList = this.fileInput.nativeElement.files;

    const accommodationData = {
      accommodation_name: name,
      accommodation_location: location,
      accommodation_amenities: amenities,
      accommodation_min_guests: minGuests,
      accommodation_max_guests: maxGuests,
    };

    // for (let i = 0; i < files.length; i++) {
    //   formData.append('images', files[i], files[i].name);
    // }

    //TODO:


    this.dataService.sendData(accommodationData).subscribe(
      (response:any) => {

        const formData = new FormData();

        for (const image of this.images) {
        formData.append('image', image, image.name);
    }

       this.uploadImages(response._id,formData);

        this.notification = { msgType: 'success', msgBody: `Successfully created accommodation;` };
        console.log('Response from server:', response);
        setTimeout(() => {
        this.router.navigate(['/home']);
        }, 1000) 
      },
      (error:any) => {
        this.notification = { msgType: 'error', msgBody: `Creating accommodation failed` };
        console.error('Error:', error);
      }
    );

    this.resetForm();
  }

  resetForm() {
    (document.getElementById('name') as HTMLInputElement).value = '';
    (document.getElementById('location') as HTMLInputElement).value = '';
    (document.getElementById('amenities') as HTMLTextAreaElement).value = '';
    (document.getElementById('minGuests') as HTMLInputElement).value = '';
    (document.getElementById('maxGuests') as HTMLInputElement).value = '';

    //this.fileInput.nativeElement.value = '';
  
  }
  getUsername() {
    return this.userService.currentUser.user.username;
  }


 uploadImages(accId: string, formData: any) {

  this.accService.uploadAccImages(accId, formData ).subscribe(
   (data: any) => {

      
    },
    (error) => {
    console.error('Error uploading images:', error);
     
    }
  );
}
}
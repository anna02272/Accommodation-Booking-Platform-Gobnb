import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Component } from '@angular/core';
import { AccommodationService } from 'src/app/services/accommodation.service';
import { AccDataService } from 'src/app/services/acc-data-service';

@Component({
  selector: 'app-create-accommodation',
  templateUrl: './create-accommodation.component.html',
  styleUrls: ['./create-accommodation.component.css']
})
export class CreateAccommodationComponent {

  //@ViewChild('fileInput') fileInput: ElementRef; // Access the file input element

  // constructor(
  //   private accService: AccommodationService
  // ) {
  // }
  constructor(private dataService: AccDataService) {}


  onSubmit() {

    const name = (document.getElementById('name') as HTMLInputElement).value;
    const location = (document.getElementById('location') as HTMLInputElement).value;
    const amenities = (document.getElementById('amenities') as HTMLTextAreaElement).value;
    const minGuests = parseInt((document.getElementById('minGuests') as HTMLInputElement).value, 10);
    const maxGuests = parseInt((document.getElementById('maxGuests') as HTMLInputElement).value, 10);

    
    //const files: FileList = this.fileInput.nativeElement.files;

    const accommodationData = {
      accommodation_name: name,
      accommodation_location: location,
      accommodation_amenities: amenities,
      accommodation_min_guests: minGuests,
      accommodation_max_guests: maxGuests,
      accommodation_image_url: 'https://www.google.com/' 
    };

    // for (let i = 0; i < files.length; i++) {
    //   formData.append('images', files[i], files[i].name);
    // }

    //TODO:


    this.dataService.sendData(accommodationData).subscribe(
      (response:any) => {
        console.log('Response from server:', response);
      },
      (error:any) => {
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
}
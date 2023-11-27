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
    const amenities = (document.getElementById('description') as HTMLTextAreaElement).value;
    const minGuests = (document.getElementById('minGuests') as HTMLInputElement).value;
    const maxGuests = (document.getElementById('maxGuests') as HTMLInputElement).value;
    //const files: FileList = this.fileInput.nativeElement.files;

    const formData = new FormData();
    formData.append('accommodation_name', name);
    formData.append('accommodation_location', location);
    formData.append('accommodation_amenities', amenities);
    formData.append('accommodation_min_guests', minGuests);
    formData.append('accommodation_max_guests', maxGuests);
    formData.append('accommodation_image_url', 'https://www.google.com/')

    // for (let i = 0; i < files.length; i++) {
    //   formData.append('images', files[i], files[i].name);
    // }

    //TODO:


    this.dataService.sendData(formData).subscribe(
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
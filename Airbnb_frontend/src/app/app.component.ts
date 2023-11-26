import { Component } from '@angular/core';
import { NgForm } from '@angular/forms';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'Gobnb';

  token: string|undefined;

  constructor() {
    this.token = undefined;
  }


   public send(form: NgForm): void {
    if (form) {
      for (const control of Object.keys(form.controls)) {
        form.controls[control].markAsTouched();
      }
    } 
    return;
  }

}


  



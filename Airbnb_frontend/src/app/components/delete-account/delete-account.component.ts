import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-delete-account',
  templateUrl: './delete-account.component.html',
  styleUrls: ['./delete-account.component.css']
})
export class DeleteAccountComponent implements OnInit {

    constructor(private router: Router) { }


ngOnInit() {
    setTimeout(() => {
      this.router.navigate(['/register']);
    }, 3000);



  }
  
}

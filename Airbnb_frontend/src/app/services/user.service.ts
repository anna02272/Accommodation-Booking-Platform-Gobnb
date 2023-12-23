import {Injectable} from '@angular/core';
import {ApiService} from './api.service';
import {ConfigService} from './config.service';
import {map} from 'rxjs/operators';
import { User } from '../models/user';
import { BehaviorSubject } from 'rxjs';
import { Observable } from 'rxjs';


@Injectable({
  providedIn: 'root'
})
export class UserService {
  currentUser: any;
  currentUserProfile: any;

  private currentUserSubject = new BehaviorSubject<User | null>(null);
  currentUser$ = this.currentUserSubject.asObservable();
  
  
  
  constructor(
    private apiService: ApiService,
    private config: ConfigService,
  ) {
  }

  getMyInfo() {
    return this.apiService.get(this.config.currentUser_url)
      .pipe(map(user => {
        this.currentUser = user;
        return user;
      }));
  }

  getProfile() {
    return this.apiService.get(this.config.currentUserProfile_url );
   }
 
   
  setCurrentUser(user: User | null) {
    this.currentUserSubject.next(user);
  }
  changePassword(current_password: string, new_password: string, confirm_new_password:string): Observable<any> {
    const changePasswordData = {
      current_password,
      new_password,
      confirm_new_password
    };
    return this.apiService.patch(this.config.changePassword_url, changePasswordData);
  }

  deleteProfile() {
  return this.apiService.delete(this.config.deleteProfile_url)
  }

}

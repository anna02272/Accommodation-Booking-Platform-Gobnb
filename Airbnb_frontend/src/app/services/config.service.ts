import {Injectable} from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {

  private _auth_api_url = 'http://localhost:8080/api';
  private _res_api_url = 'http://localhost:8082/api';
  private _acc_api_url = 'http://localhost:8083/api';
  private _profile_api_url = 'http://localhost:8084/api';
  private _rec_api_url = 'http://localhost:8085/api';
  private _notif_api_url = 'http://localhost:8086/api';


  private _auth_url = this._auth_api_url + '/auth';
  private _login_url = this._auth_url + '/login';
  private _user_url = this._auth_api_url + '/users';
  private _current_user_url = this._user_url + '/currentUser';

  get login_url(): string {
    return this._login_url;
  }

  get currentUser_url(): string {
    return this._current_user_url;
  }

  }



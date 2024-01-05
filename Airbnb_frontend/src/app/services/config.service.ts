import {Injectable} from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ConfigService {
  private _api_url = 'https://localhost:8000/api';
  // private _auth_api_url = 'https://localhost:8080/api';
  // private _acc_api_url = 'https://localhost:8083/api';
  // private _profile_api_url = 'https://localhost:8084/api';
  // private _rec_api_url = 'https://localhost:8085/api';
  // private _resv_api_url = 'https://localhost:8082/api';
  // private _rating_api_url = 'https://localhost:8087/api';
  // private _availability_api_url = 'https://localhost:8082/api';
  // private _notifications_api_url = 'https://localhost:8089/api';

  private _auth_url = this._api_url + '/auth';
  private _login_url = this._auth_url + '/login';
  private _register_url = this._auth_url + '/register';
  private _verifyEmail_url = this._auth_url + '/verifyEmail';
  private _resendVerification_url = this._auth_url + '/resendVerification';
  private _forgotPassword_url = this._auth_url + '/forgotPassword';
  private _resetPassword_url = this._auth_url + '/resetPassword';

  private _user_url = this._api_url + '/users';
  private _current_user_url = this._user_url + '/currentUser';
  private _current_user_profile_url = this._user_url + '/currentUserProfile';

  private _changePassword_url = this._user_url + '/changePassword';
  private _deleteProfile_url = this._user_url + '/delete';

  private _acc_url = this._api_url + '/accommodations';
  private _acc_delete = this._acc_url + '/delete/';
  private _host_accs = this._acc_url + '/get/host/';
  private _img_acc_all = this.acc_url + '/images/';
  private _img_acc_upload = this.acc_url + '/upload/images/';


  private _notif_url = this._api_url + '/notifications';
  private _fetchNotifications_url = this._notif_url + '/host';


  private _resv_url = this._api_url + '/reservations';
  private _create_resv_url = this._api_url + '/reservations/create';

  private _rating_url = this._api_url + '/rating';

  private _availability_url = this._api_url + 'reservations/availability';
  private _create_availability_period_url = this._api_url + 'reservations/availability/create';
  private _get_availability_url = this._api_url + '/availability/get';

  get login_url(): string {
    return this._login_url;
  }

  get register_url(): string {
    return this._register_url;
  }

  get deleteAccommodation_url(): string {
    return this._acc_delete;
  }

  get getAccommodationsByHost_url(): string {
    return this._host_accs;
  }

  get getNotificationUrl_url(): string  {
    return this._fetchNotifications_url;
  }

  get verifyEmail_url(): string {
    return this._verifyEmail_url;
  }
  get resendVerification_url(): string {
    return this._resendVerification_url;
  }
  get forgotPassword_url(): string {
    return this._forgotPassword_url;
  }
  get resetPassword_url(): string {
    return this._resetPassword_url;
  }
  get currentUser_url(): string {
    return this._current_user_url;
  }
  get currentUserProfile_url(): string {
    return this._current_user_profile_url;
  }
  get imagesUpload_url(): string {
    return this._img_acc_upload;
  }

  get imagesFetch_url(): string {
    return this._img_acc_all;
  }

  get createReservation_url(): string {
    return this._create_resv_url;
  }
  
  get changePassword_url(): string {
    return this._changePassword_url;
  }
  get acc_url(): string {
    return this._acc_url;
  }
  get resv_url(): string {
    return this._resv_url;
  }
  get deleteProfile_url(): string {
    return this._deleteProfile_url;
  }
  get rating_url(): string {
    return this._rating_url;
  }

  get availability_url(): string {
    return this._availability_url;
  }

  get createAvailabilityPeriod_url(): string {
    return this._create_availability_period_url;
  }
  get getAvailabilityPeriod_url(): string {
    return this._get_availability_url;
  }
  }



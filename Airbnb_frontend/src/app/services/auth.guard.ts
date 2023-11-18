import { Injectable } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot } from '@angular/router';
import { AuthService } from './auth.service';

@Injectable({
  providedIn: 'root'
})
export class AuthGuard implements CanActivate {

  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(route: ActivatedRouteSnapshot, _state: RouterStateSnapshot): boolean {
    if (this.authService.tokenIsPresent()) {
      const roles = route.data['roles'] as string[];
      if (roles && roles.length > 0) {
        const userRole = this.authService.getRole();

        if (roles.includes(userRole)) {
          return true;
        } else {
          this.router.navigate(['/home']);
          return false;
        }
      }
      return true;
    } else {
      this.router.navigate(['/home']);
      return false;
    }
  }
}

import { Address } from "./address";
import { Gender } from "./gender";
import { UserRole } from "./userRole";

export class User {
  id: number;
  username: string;
  password: string;
  email: string;
  name: string;
  lastname: string;
  address: Address;
  age: number;
  gender: Gender;
  userRole: UserRole;
  

  constructor(id: number, username: string, password: string, email:string, name: string, lastname: string, address: Address, age: number, gender: Gender, userRole: UserRole) {
    this.id = id;
    this.username = username;
    this.password = password;
    this.email = email;
    this.name = name;
    this.lastname = lastname;
    this.address = address;
    this.age = age;
    this.gender = gender;
    this.userRole = userRole;

   
  }
}
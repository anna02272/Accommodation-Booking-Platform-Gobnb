export class Reservation {

    accommodation_id: string;
    accommodation_name: string;
    accommodation_location: string;
    check_in_date!: string;
    check_out_date: string;

     constructor(accommodation_id: string, accommodation_name: string,
         accommodation_location: string, check_in_date: any, check_out_date: any) {
      this.accommodation_id = accommodation_id;
      this.accommodation_name = accommodation_name;
      this.accommodation_location = accommodation_location;
      this.check_in_date = check_in_date,
      this.check_out_date = check_out_date

    }


}
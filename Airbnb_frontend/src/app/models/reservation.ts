export class Reservation {

    accommodation_id: string;
    check_in_date!: string;
    check_out_date: string;
    number_of_guests: number;

     constructor(accommodation_id: string,check_in_date: any, check_out_date: any, number_of_guests: number) {
      this.accommodation_id = accommodation_id;
      this.check_in_date = check_in_date,
      this.check_out_date = check_out_date
      this.number_of_guests = number_of_guests

    }


}
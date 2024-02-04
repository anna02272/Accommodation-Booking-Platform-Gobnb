export class GetReservation {
  reservation_id_time_created: string;
  guest_id: string;
  accommodation_id: string;
    accommodation_name: string;
    accommodation_location: string;
    check_in_date: Date;
    check_out_date: Date;
    number_of_guests: number
   
    constructor(
        reservation_id_time_created: string,
        guest_id: string,
        accommodation_id: string, 
        accommodation_name: string,
        accommodation_location: string,
        check_in_date: Date,
        check_out_date: Date,
        number_of_guests: number
 
    ) {
      this.reservation_id_time_created = reservation_id_time_created;
      this.guest_id = guest_id;
      this.accommodation_id = accommodation_id;
      this.accommodation_name = accommodation_name;
      this.accommodation_location = accommodation_location;
      this.check_in_date = check_in_date;
      this.check_out_date = check_out_date;
      this.number_of_guests = number_of_guests;
   
    }
  }
  
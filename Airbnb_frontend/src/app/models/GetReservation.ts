export class GetReservation {
    ReservationIdTimeCreated: string;
    GuestId: string;
    AccommodationId: string;
    AccommodationName: string;
    AccommodationLocation: string;
    CheckInDate: Date;
    CheckOutDate: Date;
    NumberOfGuests: number
   
    constructor(
        ReservationIdTimeCreated: string,
        GuestId: string,
        AccommodationId: string, 
        AccommodationName: string,
        AccommodationLocation: string,
        CheckInDate: Date,
        CheckOutDate: Date,
        NumberOfGuests: number
 
    ) {
      this.ReservationIdTimeCreated = ReservationIdTimeCreated;
      this.GuestId = GuestId;
      this.AccommodationId = AccommodationId;
      this.AccommodationName = AccommodationName;
      this.AccommodationLocation = AccommodationLocation;
      this.CheckInDate = CheckInDate;
      this.CheckOutDate = CheckOutDate;
      this.NumberOfGuests = NumberOfGuests;
   
    }
  }
  
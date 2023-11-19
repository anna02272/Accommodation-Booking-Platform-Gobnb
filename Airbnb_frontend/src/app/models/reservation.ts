export class Reservation {

    guestId: string;
    accommodatioonId: string;
    accommodationName: string;
    accommodationLocation: string;
    checkInDate: Date;
    CheckOutDate: Date;

     constructor(guestId: string, accommodatioonId: string, accommodationName: string,
         accommodationLocation: string, checkInDate: Date, CheckOutDate: Date) {
      this.guestId = guestId;
      this.accommodatioonId = accommodatioonId;
      this.accommodationName = accommodationName;
      this.accommodationLocation = accommodationLocation;
      this.checkInDate = checkInDate,
      this.CheckOutDate = CheckOutDate

    }


}
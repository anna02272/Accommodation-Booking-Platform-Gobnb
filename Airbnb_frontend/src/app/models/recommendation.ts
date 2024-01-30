export class Recommendation {
    accommodation_id: string;
    accommodation_name: string;
    accommodation_location: string;
    accommodation_min_guests: number;
    accommodation_max_guests: number;
    // flagCanDelete?: boolean;
    // coverImage?: string;


    constructor(
      accommodation_id: string,
      accommodation_name: string,
      accommodation_location: string,
    //   accommodation_amenities: Map<string, boolean>,
      accommodation_min_guests: number,
      accommodation_max_guests: number,

    ) {
      this.accommodation_id = accommodation_id;
      this.accommodation_name = accommodation_name;
      this.accommodation_location = accommodation_location;
      this.accommodation_min_guests = accommodation_min_guests;
      this.accommodation_max_guests = accommodation_max_guests;

    }

  }
// export class Recommendation {
//     accommodation: {
//       _id: string;
//       accommodation_name: string;
//       accommodation_location: string;
//       accommodation_min_guests: number;
//       accommodation_max_guests: number;
//       accommodation_active: boolean;
//       host_id: string;
//     };
  
//     constructor(accommodation: {
//       _id: string;
//       accommodation_name: string;
//       accommodation_location: string;
//       accommodation_min_guests: number;
//       accommodation_max_guests: number;
//       accommodation_active: boolean;
//       host_id: string;
//     }) {
//       this.accommodation = accommodation;
//     }
//   }
  

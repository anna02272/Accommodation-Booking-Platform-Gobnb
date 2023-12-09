export class Accommodation {
    host_id: string; 
    accommodation_name: string;
    accommodation_location: string;
    accommodation_id: string; 
    accommodation_amenities: string;
    accommodation_min_guests: number;
    accommodation_max_guests: number;
    accommodation_image_url: string;

  
    constructor(
      host_id: string,
      accommodation_name: string,
      accommodation_location: string,
      accommodation_id: string,
      accommodation_amenities: string,
      accommodation_min_guests: number,
      accommodation_max_guests: number,
      accommodation_image_url: string,

    ) {
      this.host_id = host_id;
      this.accommodation_name = accommodation_name;
      this.accommodation_location = accommodation_location;
      this.accommodation_id = accommodation_id;
      this.accommodation_amenities = accommodation_amenities;
      this.accommodation_min_guests = accommodation_min_guests;
      this.accommodation_max_guests = accommodation_max_guests;
      this.accommodation_image_url = accommodation_image_url;
    }
  }
  
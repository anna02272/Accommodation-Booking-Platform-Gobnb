export class Accommodation {
    host_id: string; 
    accommodation_name: string;
    accommodation_location: string;
    _id: string; 
    accommodation_amenities: Map<string, boolean>;
    accommodation_min_guests: number;
    accommodation_max_guests: number;
    accommodation_images: Array<string>;

  
    constructor(
      host_id: string,
      accommodation_name: string,
      accommodation_location: string,
      _id: string,
      accommodation_amenities: Map<string, boolean>,
      accommodation_min_guests: number,
      accommodation_max_guests: number,
      accommodation_images: Array<string>

    ) {
      this.host_id = host_id;
      this.accommodation_name = accommodation_name;
      this.accommodation_location = accommodation_location;
      this._id = _id;
      this.accommodation_amenities = accommodation_amenities;
      this.accommodation_min_guests = accommodation_min_guests;
      this.accommodation_max_guests = accommodation_max_guests;
      this.accommodation_images = accommodation_images;
    }
  }
  
import { Rating } from "./rating";
import { User } from "./user";
import { Accommodation } from "./accommodation";


export class RateAccommodation {
  averageRating: number;
  ratings: RatingItem[];

  constructor(averageRating: number, ratings: RatingItem[]) {
    this.averageRating = averageRating;
    this.ratings = ratings;
  }
}

export class RatingItem {
  id: string;
  accommodation: Accommodation;
  guest: User;
  dateAndTime: Date;
  rating: number;

  constructor(id: string, accommodation: Accommodation, guest: User, dateAndTime: Date, rating: number) {
    this.id = id;
    this.accommodation = accommodation;
    this.guest = guest;
    this.dateAndTime = dateAndTime;
    this.rating = rating;
  }
}

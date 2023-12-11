import { Rating } from "./rating";
import { User } from "./user";

export class RateHost {
  averageRating: number;
  ratings: RatingItem[];

  constructor(averageRating: number, ratings: RatingItem[]) {
    this.averageRating = averageRating;
    this.ratings = ratings;
  }
}

export class RatingItem {
  id: string;
  host: User;
  guest: User;
  dateAndTime: Date;
  rating: number;

  constructor(id: string, host: User, guest: User, dateAndTime: Date, rating: number) {
    this.id = id;
    this.host = host;
    this.guest = guest;
    this.dateAndTime = dateAndTime;
    this.rating = rating;
  }
}

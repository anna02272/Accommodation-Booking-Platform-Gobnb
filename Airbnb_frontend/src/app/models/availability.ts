export class Availability {
    date: Date;
    Price: number;
    PriceType: string;
    AvailabilityType: string;


    constructor(
        date: Date,
        Price: number,
        PriceType: string,
        AvailabilityType: string
    ) {
        this.date = date;
        this.Price = Price;
        this.PriceType = PriceType;
        this.AvailabilityType = AvailabilityType;
    }
}

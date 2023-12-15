export class AvailabilityPeriod {
    start_date: Date;
    end_date: Date;
    Price: number;
    PriceType: string;
    AvailabilityType: string;


    constructor(
        start_date: Date,
        end_date: Date,
        Price: number,
        PriceType: string,
        AvailabilityType: string
    ) {
        this.start_date = start_date;
        this.end_date = end_date;
        this.Price = Price;
        this.PriceType = PriceType;
        this.AvailabilityType = AvailabilityType;
    }
}

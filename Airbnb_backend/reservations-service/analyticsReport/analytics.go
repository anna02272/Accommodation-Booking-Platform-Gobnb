package analyticsReport

import (
	"context"
	"fmt"
	ga "google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
	"strconv"
)

func GetPropertiesAnalytics() {
	//fmt.Println("GET PROPERITES ANALYTICS HERE")
	ctx := context.Background()
	client, err := ga.NewService(ctx, option.WithCredentialsFile("/app/gobnb-409715-26445f8b186e.json"))

	if err != nil {
		panic(err)
	}

	runReportRequest := &ga.RunReportRequest{
		DateRanges: []*ga.DateRange{
			{
				StartDate: "30daysAgo",
				EndDate:   "today",
			},
		},

		Metrics: []*ga.Metric{
			{
				Name: "active1DayUsers",
			},
		},
	}

	r, err := client.Properties.RunReport("properties/421890861", runReportRequest).Do()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Rows, "Rows")
	fmt.Println(r.RowCount, "CountR")
}

func GetNumberOfUsersPerPage(fromDate string, toDate string, pageName string, metrics []*ga.Metric) (float64, float64) {
	ctx := context.Background()
	client, err := ga.NewService(ctx, option.WithCredentialsFile("/app/gobnb-409715-26445f8b186e.json"))
	if err != nil {
		panic(err)
	}

	runReportRequest := &ga.RunReportRequest{
		DateRanges: []*ga.DateRange{
			{
				StartDate: fromDate,
				EndDate:   toDate,
			},
		},

		Metrics: metrics,
		Dimensions: []*ga.Dimension{
			{
				Name: "pagePath",
			},
		},
		DimensionFilter: &ga.FilterExpression{
			Filter: &ga.Filter{
				FieldName: "pagePath",
				StringFilter: &ga.StringFilter{
					MatchType: "EXACT",
					Value:     pageName,
				},
			},
		},
	}

	r, err := client.Properties.RunReport("properties/421890861", runReportRequest).Do()
	if err != nil {
		fmt.Println(err)
	}

	var result []float64

	if r.RowCount > 0 {
		for _, value := range r.Rows {
			for _, metricValue := range value.MetricValues {
				floatValue, err := strconv.ParseFloat(metricValue.Value, 64)
				if err != nil {
					fmt.Println(err)
				}
				result = append(result, floatValue)
			}
		}
	}

	return result[0], result[1]

}

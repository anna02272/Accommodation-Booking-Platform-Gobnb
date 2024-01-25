package repository

import (
	"context"
	"fmt"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"reservations-service/data"
	"time"
)

type ReportRepo struct {
	session *gocql.Session //connection towards CassandraDB
	logger  *log.Logger
	ctx     context.Context
	Tracer  trace.Tracer
}

func NewReportRepo(logger *log.Logger, tracer trace.Tracer) (*ReportRepo, error) {
	db := os.Getenv("CASS_DB")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	// Create 'report' keyspace
	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "report", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to reservation keyspace
	cluster.Keyspace = "report"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &ReportRepo{
		session: session,
		logger:  logger,
		Tracer:  tracer,
	}, nil
}

// Disconnect from database
func (sr *ReportRepo) CloseSessionEventReport() {
	sr.session.Close()
}

func (sr *ReportRepo) CreateTableMonthlyReport() {
	err := sr.session.Query(
		`CREATE TABLE IF NOT EXISTS monthly_report (
        report_id_time_created timeuuid,
        accommodation_id text,
        year int,
        month int,
        reservation_count int,
        rating_count int,
        page_visits int,
        avg_visit_time double,
        PRIMARY KEY ((accommodation_id, year, month), report_id_time_created)
    ) WITH CLUSTERING ORDER BY (report_id_time_created DESC);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
	if err != nil {
		//span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
	}
}

func (sr *ReportRepo) CreateTableDailyReport() {
	err := sr.session.Query(
		`CREATE TABLE IF NOT EXISTS daily_report (
        report_id_time_created timeuuid,
        accommodation_id text,
        date_created timestamp,
        reservation_count int,
        rating_count int,
        page_visits int,
        avg_visit_time double,
        PRIMARY KEY ((accommodation_id, date_created), report_id_time_created)
    ) WITH CLUSTERING ORDER BY (report_id_time_created DESC);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
	if err != nil {
		//span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
	}
}

func (sr *ReportRepo) InsertDailyReport(ctx context.Context, dailyReport *data.DailyReport) error {
	ctx, span := sr.Tracer.Start(ctx, "ReportRepository.InsertDailyReport")
	defer span.End()

	reportIdTimeCreated := gocql.TimeUUID()
	dateCreated := time.Now()
	startOfDay := dateCreated.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)
	fmt.Println(startOfDay)
	fmt.Println("start of the day")
	fmt.Println(endOfDay)
	fmt.Println("end of the day")

	var existingReportId gocql.UUID
	var dateCreatedFromDB time.Time
	if err := sr.session.Query(
		`SELECT report_id_time_created, date_created FROM daily_report 
         WHERE accommodation_id = ? AND date_created >= ? AND date_created <= ? LIMIT 1 ALLOW FILTERING`,
		dailyReport.AccommodationID,
		startOfDay,
		endOfDay,
	).WithContext(ctx).Scan(&existingReportId, &dateCreatedFromDB); err == nil {
		fmt.Println(existingReportId)
		fmt.Println("found existing report id")
		err := sr.session.Query(
			`UPDATE daily_report SET 
             reservation_count = ?, rating_count = ?, page_visits = ?, avg_visit_time = ? 
             WHERE report_id_time_created = ? AND date_created = ? AND accommodation_id = ?`,
			dailyReport.ReservationCount,
			dailyReport.RatingCount,
			dailyReport.PageVisits,
			dailyReport.AverageVisitTime,
			existingReportId,
			dateCreatedFromDB,
			dailyReport.AccommodationID,
		).WithContext(ctx).Exec()

		fmt.Println(existingReportId)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			sr.logger.Println(err)
			fmt.Println(err)
			fmt.Println("float error here")
			return err
		}
	} else {
		// Insert a new daily report if it doesn't exist
		err := sr.session.Query(
			`INSERT INTO daily_report 
             (report_id_time_created,accommodation_id,date_created,reservation_count,rating_count,
              page_visits,avg_visit_time) 
             VALUES (?, ?, ?, ?, ?, ?, ?)`,
			reportIdTimeCreated,
			dailyReport.AccommodationID,
			dateCreated,
			dailyReport.ReservationCount,
			dailyReport.RatingCount,
			dailyReport.PageVisits,
			dailyReport.AverageVisitTime,
		).WithContext(ctx).Exec()

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			sr.logger.Println(err)
			return err
		}
	}

	return nil
}

func (sr *ReportRepo) InsertMonthlyReport(ctx context.Context, monthlyReport *data.MonthlyReport) error {
	ctx, span := sr.Tracer.Start(ctx, "ReportRepository.InsertMonthlyReport")
	defer span.End()

	reportIdTimeCreated := gocql.TimeUUID()
	currentTime := time.Now().UTC()
	currentYear, currentMonth, _ := currentTime.Date()

	var existingReportId gocql.UUID
	if err := sr.session.Query(
		`SELECT report_id_time_created FROM monthly_report 
         WHERE accommodation_id = ? AND year = ? AND month = ? LIMIT 1 ALLOW FILTERING`,
		monthlyReport.AccommodationID,
		currentYear,
		int(currentMonth),
	).WithContext(ctx).Scan(&existingReportId); err == nil {
		fmt.Println("Found report")
		fmt.Println(existingReportId)
		err := sr.session.Query(
			`UPDATE monthly_report SET 
             reservation_count = ?, rating_count = ?, page_visits = ?, avg_visit_time = ? 
             WHERE report_id_time_created = ? AND accommodation_id = ? AND year = ? AND month = ?`,
			monthlyReport.ReservationCount,
			monthlyReport.RatingCount,
			monthlyReport.PageVisits,
			monthlyReport.AverageVisitTime,
			existingReportId,
			monthlyReport.AccommodationID,
			currentYear,
			int(currentMonth),
		).WithContext(ctx).Exec()

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			sr.logger.Println(err)
			fmt.Println(err)
			fmt.Println("error updating existing monthly report")
			return err
		}
	} else {
		// Insert a new monthly report if it doesn't exist
		err := sr.session.Query(
			`INSERT INTO monthly_report 
             (report_id_time_created,accommodation_id,year,month,reservation_count,rating_count,
              page_visits,avg_visit_time) 
             VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			reportIdTimeCreated,
			monthlyReport.AccommodationID,
			currentYear,
			int(currentMonth),
			monthlyReport.ReservationCount,
			monthlyReport.RatingCount,
			monthlyReport.PageVisits,
			monthlyReport.AverageVisitTime,
		).WithContext(ctx).Exec()

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			sr.logger.Println(err)
			fmt.Println(err)
			fmt.Println("error inserting new monthly report")
			return err
		}
	}

	return nil
}

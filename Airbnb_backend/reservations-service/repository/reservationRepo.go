package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"reservations-service/data"
	"time"
)

type ReservationRepo struct {
	session *gocql.Session //connection towards CassandraDB
	logger  *log.Logger
	ctx     context.Context
	Tracer  trace.Tracer
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
// if CassandrDB exists, this function connects to DB,if not it tries to create cassandraDB
func New(logger *log.Logger, tracer trace.Tracer) (*ReservationRepo, error) {
	db := os.Getenv("CASS_DB")

	cluster := gocql.NewCluster(db)
	cluster.Keyspace = "system"
	session, err := cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	// Create 'reservation' keyspace
	err = session.Query(
		fmt.Sprintf(`CREATE KEYSPACE IF NOT EXISTS %s
					WITH replication = {
						'class' : 'SimpleStrategy',
						'replication_factor' : %d
					}`, "reservation", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to reservation keyspace
	cluster.Keyspace = "reservation"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &ReservationRepo{
		session: session,
		logger:  logger,
		Tracer:  tracer,
	}, nil
}

// Disconnect from database
func (sr *ReservationRepo) CloseSession() {
	sr.session.Close()
}

// Create reservations_by_guest table
func (sr *ReservationRepo) CreateTable() {
	//ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.CreateTable")
	//defer span.End()

	err := sr.session.Query(
		`CREATE TABLE IF NOT EXISTS reservations_by_guest (
        reservation_id_time_created timeuuid,
        guest_id text,
        accommodation_id text,
        accommodation_name text,
        accommodation_location text,
        accommodation_host_id text,
        check_in_date timestamp,
        check_out_date timestamp,
        number_of_guests int,
        isCanceled boolean ,
        PRIMARY KEY ((guest_id, reservation_id_time_created),check_in_date)
    ) WITH CLUSTERING ORDER BY (check_in_date ASC);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}

	err = sr.session.Query(
		`CREATE INDEX IF NOT EXISTS idx_accommodation_id ON reservations_by_guest (accommodation_id);`,
	).Exec()

	err = sr.session.Query(
		`CREATE INDEX IF NOT EXISTS idx_check_in_date ON reservations_by_guest (check_in_date);`,
	).Exec()

	if err != nil {
		//span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
	}
}

func (sr *ReservationRepo) InsertReservationByGuest(ctx context.Context, guestReservation *data.ReservationByGuestCreate,
	guestId string, accommodationName string, accommodationLocation string, accommodationHostId string) error {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.InsertReservationByGuest")
	defer span.End()

	// Check if there is an existing reservation for the same guest, accommodation, and check-in date
	var existingReservationCount int
	errSameReservation := sr.session.Query(
		`SELECT COUNT(*) FROM reservations_by_guest 
         WHERE guest_id = ? AND accommodation_id = ? AND check_in_date = ? AND isCanceled = false ALLOW FILTERING`,
		guestId, guestReservation.AccommodationId, guestReservation.CheckInDate,
	).WithContext(ctx).Scan(&existingReservationCount)

	if errSameReservation != nil {
		span.SetStatus(codes.Error, errSameReservation.Error())
		sr.logger.Println(errSameReservation)
		return errSameReservation
	}

	if existingReservationCount > 0 {
		span.SetStatus(codes.Error, "Guest already has a reservation for the same accommodation and check-in date")
		return fmt.Errorf("Guest already has a reservation for the same accommodation and check-in date")
	}

	fmt.Println(errSameReservation)

	// If no existing reservation is found, proceed with the insertion
	reservationIdTimeCreated := gocql.TimeUUID()

	err := sr.session.Query(
		`INSERT INTO reservations_by_guest 
         (reservation_id_time_created, guest_id,accommodation_id, accommodation_name,accommodation_location, accommodation_host_id,
          check_in_date, check_out_date,number_of_guests, isCanceled) 
         VALUES (?, ?, ?, ?, ?, ?, ?,?, ?, false)`,
		reservationIdTimeCreated,
		guestId,
		guestReservation.AccommodationId,
		accommodationName,
		accommodationLocation,
		accommodationHostId,
		guestReservation.CheckInDate,
		guestReservation.CheckOutDate,
		guestReservation.NumberOfGuests,
	).WithContext(ctx).Exec()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
		return err
	}

	return nil
}

func (sr *ReservationRepo) GetAllReservations(ctx context.Context, guestID string) (data.ReservationsByGuest, error) {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.GetAllReservations")
	defer span.End()

	query := `SELECT  reservation_id_time_created, guest_id, accommodation_id,
        accommodation_location, accommodation_host_id, accommodation_name, check_in_date, check_out_date, number_of_guests FROM reservation.reservations_by_guest WHERE guest_id = ?  AND isCanceled = false ALLOW FILTERING`

	//iterable := sr.session.Query(query, guestID).Iter()
	iterable := sr.session.Query(query, guestID).WithContext(ctx).Iter()

	var reservations data.ReservationsByGuest
	m := map[string]interface{}{}

	for iterable.MapScan(m) {
		res := data.ReservationByGuest{

			ReservationIdTimeCreated: data.TimeUUID(m["reservation_id_time_created"].(gocql.UUID)),
			GuestId:                  m["guest_id"].(string),
			AccommodationId:          m["accommodation_id"].(string),
			AccommodationLocation:    m["accommodation_location"].(string),
			AccommodationName:        m["accommodation_name"].(string),
			AccommodationHostId:      m["accommodation_host_id"].(string),
			CheckInDate:              m["check_in_date"].(time.Time),
			CheckOutDate:             m["check_out_date"].(time.Time),
			NumberOfGuests:           m["number_of_guests"].(int),
		}

		reservations = append(reservations, &res)
		m = map[string]interface{}{}
	}

	if err := iterable.Close(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
		return nil, err
	}

	return reservations, nil
}

func (sr *ReservationRepo) GetReservationByAccommodationIDAndCheckOut(ctx context.Context, accommodationId string) int {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.GetReservationByAccommodationIDAndCheckOut")
	defer span.End()

	var countVariable int
	var checkOutNow = time.Now()

	query := `
		SELECT COUNT(*) FROM reservations_by_guest 
         WHERE check_out_date >= ? AND accommodation_id = ?  AND isCanceled = false ALLOW FILTERING`

	if err := sr.session.Query(query, checkOutNow, accommodationId).WithContext(ctx).Scan(&countVariable); err != nil {
		span.SetStatus(codes.Error, "Error retrieving reservation details: "+err.Error())
		sr.logger.Println("Error retrieving reservation details:" + err.Error())
		return -1
	}
	return countVariable

}

func (sr *ReservationRepo) GetReservationAccommodationID(ctx context.Context, reservationID string, guestID string) (string, error) {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.GetReservationAccommodationID")
	defer span.End()

	var accommodationID string
	query := `
		SELECT accommodation_id FROM reservations_by_guest
         WHERE guest_id = ? AND reservation_id_time_created = ?  AND isCanceled = false ALLOW FILTERING`
	fmt.Println(reservationID)
	fmt.Println("repo rsv id")
	if err := sr.session.Query(query, guestID, reservationID).Scan(&accommodationID); err != nil {
		span.SetStatus(codes.Error, "Error retrieving reservation details: "+err.Error())
		sr.logger.Println("Error retrieving reservation details:", err)
		return "", err
	}
	return accommodationID, nil

}

func (sr *ReservationRepo) CancelReservationByID(ctx context.Context, guestID string, reservationID string, checkInDate time.Time) error {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.CancelReservationByID")
	defer span.End()

	var isCanceled bool
	checkQuery := `SELECT isCanceled FROM reservations_by_guest 
        WHERE guest_id = ? AND reservation_id_time_created = ? AND check_in_date = ?`
	if err := sr.session.Query(checkQuery, guestID, reservationID, checkInDate).WithContext(ctx).Scan(&isCanceled); err != nil {
		span.SetStatus(codes.Error, "Error checking reservation status: "+err.Error())
		sr.logger.Println("Error checking reservation status:", err)
		return err
	}

	if isCanceled {
		return errors.New("reservation is already canceled")
	}

	updateQuery := `UPDATE reservations_by_guest SET isCanceled = true
        WHERE guest_id = ? AND reservation_id_time_created = ? AND check_in_date = ?`

	if err := sr.session.Query(updateQuery, guestID, reservationID, checkInDate).WithContext(ctx).Exec(); err != nil {
		span.SetStatus(codes.Error, "Error canceling reservation: "+err.Error())
		sr.logger.Println("Error canceling reservation:", err)
		return err
	}

	return nil
}

func (sr *ReservationRepo) GetReservationCheckInDate(ctx context.Context, reservationID string, guestID string) (time.Time, error) {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.GetReservationCheckInDate")
	defer span.End()

	var checkInDate time.Time

	query := `
        SELECT check_in_date FROM reservations_by_guest
        WHERE guest_id = ? AND reservation_id_time_created = ?`

	if err := sr.session.Query(query, guestID, reservationID).WithContext(ctx).Scan(&checkInDate); err != nil {
		span.SetStatus(codes.Error, "Error retrieving check-in date: "+err.Error())
		sr.logger.Println("Error retrieving check-in date:", err)
		return time.Time{}, err
	}

	return checkInDate, nil
}
func (sr *ReservationRepo) GetReservationCheckOutDate(ctx context.Context, reservationID string, guestID string) (time.Time, error) {
	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.GetReservationCheckOutDate")
	defer span.End()

	var checkOutDate time.Time

	query := `
        SELECT check_out_date FROM reservations_by_guest
        WHERE guest_id = ? AND reservation_id_time_created = ?`

	if err := sr.session.Query(query, guestID, reservationID).WithContext(ctx).Scan(&checkOutDate); err != nil {
		span.SetStatus(codes.Error, "Error retrieving check-out date: "+err.Error())
		sr.logger.Println("Error retrieving check-out date:", err)
		return time.Time{}, err
	}

	return checkOutDate, nil
}

//func (sr *ReservationRepo) CancelReservationByID(ctx context.Context, guestID string, reservationID string) error {
//	ctx, span := sr.Tracer.Start(ctx, "ReservationRepository.CancelReservationByID")
//	defer span.End()
//	var checkInDate time.Time
//	query := `
//        SELECT check_in_date
//        FROM reservation.reservations_by_guest
//        WHERE guest_id = ? AND reservation_id_time_created = ?`
//
//	if err := sr.session.Query(query, guestID, reservationID).WithContext(ctx).Scan(&checkInDate); err != nil {
//		span.SetStatus(codes.Error, "Error retrieving reservation details: "+err.Error())
//		sr.logger.Println("Error retrieving reservation details:", err)
//		return err
//	}
//
//	currentTime := time.Now()
//	if currentTime.After(checkInDate) {
//		span.SetStatus(codes.Error, "cannot cancel reservation, check-in date has already started")
//		return errors.New("cannot cancel reservation, check-in date has already started")
//	}
//
//	deleteQuery := ` DELETE FROM reservations_by_guest
//        WHERE guest_id = ? AND reservation_id_time_created = ?`
//
//	if err := sr.session.Query(deleteQuery, guestID, reservationID).WithContext(ctx).Exec(); err != nil {
//		span.SetStatus(codes.Error, "Error canceling reservation: "+err.Error())
//		sr.logger.Println("Error canceling reservation:", err)
//		return err
//	}
//
//	return nil
//}

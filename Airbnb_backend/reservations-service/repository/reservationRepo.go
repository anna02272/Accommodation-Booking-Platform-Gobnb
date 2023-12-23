package repository

import (
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"os"
	"reservations-service/data"
	"time"
)

type ReservationRepo struct {
	session *gocql.Session
	logger  *log.Logger
}

func New(logger *log.Logger) (*ReservationRepo, error) {
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
	}, nil
}

// Disconnect from database
func (sr *ReservationRepo) CloseSession() {
	sr.session.Close()
}

// Create reservations_by_guest table
func (sr *ReservationRepo) CreateTable() {

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
		`CREATE INDEX IF NOT EXISTS idx_check_out_date ON reservations_by_guest (check_out_date);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
}

func (sr *ReservationRepo) InsertReservationByGuest(guestReservation *data.ReservationByGuestCreate,
	guestId string, accommodationName string, accommodationLocation string, accommodationHostId string) error {
	// Check if there is an existing reservation for the same guest, accommodation, and check-in date
	var existingReservationCount int
	errSameReservation := sr.session.Query(
		`SELECT COUNT(*) FROM reservations_by_guest 
         WHERE guest_id = ? AND accommodation_id = ? AND check_in_date = ? ALLOW FILTERING`,
		guestId, guestReservation.AccommodationId, guestReservation.CheckInDate,
	).Scan(&existingReservationCount)

	if errSameReservation != nil {
		sr.logger.Println(errSameReservation)
		return errSameReservation
	}

	if existingReservationCount > 0 {
		return fmt.Errorf("Guest already has a reservation for the same accommodation and check-in date")
	}

	fmt.Println(errSameReservation)

	// If no existing reservation is found, proceed with the insertion
	reservationIdTimeCreated := gocql.TimeUUID()

	err := sr.session.Query(
		`INSERT INTO reservations_by_guest 
         (reservation_id_time_created, guest_id,accommodation_id, accommodation_name,accommodation_location, accommodation_host_id,
          check_in_date, check_out_date,number_of_guests) 
         VALUES (?, ?, ?, ?, ?, ?, ?,?, ?)`,
		reservationIdTimeCreated,
		guestId,
		guestReservation.AccommodationId,
		accommodationName,
		accommodationLocation,
		accommodationHostId,
		guestReservation.CheckInDate,
		guestReservation.CheckOutDate,
		guestReservation.NumberOfGuests,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
		return err
	}

	return nil
}

func (sr *ReservationRepo) GetAllReservations(guestID string) (data.ReservationsByGuest, error) {
	query := `SELECT  reservation_id_time_created, guest_id, accommodation_id,
        accommodation_location, accommodation_host_id, accommodation_name, check_in_date, check_out_date, number_of_guests FROM reservation.reservations_by_guest WHERE guest_id = ? ALLOW FILTERING`

	iterable := sr.session.Query(query, guestID).Iter()

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
		sr.logger.Println(err)
		return nil, err
	}

	return reservations, nil
}

func (sr *ReservationRepo) GetReservationByAccommodationIDAndCheckOut(accommodationId string) int {
	var countVariable int
	var checkOutNow = time.Now()

	query := `
		SELECT COUNT(*) FROM reservations_by_guest 
         WHERE check_out_date >= ? AND accommodation_id = ? ALLOW FILTERING`

	if err := sr.session.Query(query, checkOutNow, accommodationId).Scan(&countVariable); err != nil {
		sr.logger.Println("Error retrieving reservation details:", err)
		return -1
	}
	return countVariable

}

func (sr *ReservationRepo) GetReservationAccommodationID(reservationID string, guestID string) (string, error) {
	var accommodationID string
	query := `
		SELECT accommodation_id FROM reservations_by_guest
         WHERE guest_id = ? AND reservation_id_time_created = ?`
	fmt.Println(reservationID)
	fmt.Println("repo rsv id")
	if err := sr.session.Query(query, guestID, reservationID).Scan(&accommodationID); err != nil {
		sr.logger.Println("Error retrieving reservation details:", err)
		return "", err
	}
	return accommodationID, nil

}

func (sr *ReservationRepo) CancelReservationByID(guestID string, reservationID string) error {
	var checkInDate time.Time
	query := `
        SELECT check_in_date
        FROM reservation.reservations_by_guest
        WHERE guest_id = ? AND reservation_id_time_created = ?`

	if err := sr.session.Query(query, guestID, reservationID).Scan(&checkInDate); err != nil {
		sr.logger.Println("Error retrieving reservation details:", err)
		return err
	}

	currentTime := time.Now()
	if currentTime.After(checkInDate) {
		return errors.New("cannot cancel reservation, check-in date has already started")
	}

	deleteQuery := ` DELETE FROM reservations_by_guest 
        WHERE guest_id = ? AND reservation_id_time_created = ?`

	if err := sr.session.Query(deleteQuery, guestID, reservationID).Exec(); err != nil {
		sr.logger.Println("Error canceling reservation:", err)
		return err
	}

	return nil
}

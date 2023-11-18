package repository

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"os"
	"reservations-service/data"
)

// NoSQL: ReservationRepo struct encapsulating Cassandra api client
type ReservationRepo struct {
	session *gocql.Session //connection towards CassandraDB
	logger  *log.Logger    //write messages inside Console
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
// if CassandrDB exists, this function connects to DB,if not it tries to create cassandraDB
func New(logger *log.Logger) (*ReservationRepo, error) {
	db := os.Getenv("CASS_DB")

	// Connect to default keyspace
	//keyspace -something like schema in RDBMS, similar tables are in one keyspace, logical group of tables
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
        accommodation_id UUID,
        accommodation_name text,
        accommodation_location text,
        check_in_date timestamp,
        check_out_date timestamp,
        PRIMARY KEY ((guest_id, reservation_id_time_created),check_in_date)
    ) WITH CLUSTERING ORDER BY (check_in_date ASC);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}

	err = sr.session.Query(
		`CREATE INDEX IF NOT EXISTS idx_accommodation_id ON reservations_by_guest (accommodation_id);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
}

// inserting reservation into table reservation_by_guest
func (sr *ReservationRepo) InsertReservationByGuest(guestReservation *data.ReservationByGuestCreate, guestId string) error {
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
         (reservation_id_time_created, guest_id,accommodation_id, accommodation_name,accommodation_location, check_in_date, check_out_date) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		reservationIdTimeCreated,
		guestId,
		guestReservation.AccommodationId,
		guestReservation.AccommodationName,
		guestReservation.AccommodationLocation,
		guestReservation.CheckInDate,
		guestReservation.CheckOutDate,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
		return err
	}

	return nil
}

func (sr *ReservationRepo) GetReservationsByGuest(id string) (data.ReservationsByGuest, error) {
	scanner := sr.session.Query(`SELECT reservation_id_time_created, guest_id, accommodation_id, 
       accommodation_name, accommodation_location,
      check_in_date, check_out_date
FROM reservations_by_guest WHERE guest_id = ?`,
		id).Iter().Scanner()

	var reservations data.ReservationsByGuest
	for scanner.Next() {
		var rsv data.ReservationByGuest
		err := scanner.Scan(&rsv.ReservationIdTimeCreated, &rsv.GuestId,
			&rsv.AccommodationId,
			&rsv.AccommodationName, &rsv.AccommodationLocation,
			&rsv.CheckInDate, &rsv.CheckOutDate)
		if err != nil {
			sr.logger.Println(err)
			return nil, err
		}
		reservations = append(reservations, &rsv)
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return nil, err
	}
	return reservations, nil
}

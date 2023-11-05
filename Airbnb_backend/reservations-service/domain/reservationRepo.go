package domain

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"os"
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
        reservation_id UUID,
        guest_id UUID,
        accommodation_id UUID,
        accommodation_name text,
        accommodation_location text,
        check_in_date timestamp,
        check_out_date timestamp,
        PRIMARY KEY (reservation_id, guest_id)
    ) WITH CLUSTERING ORDER BY (guest_id ASC);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
}

// inserting reservation into table reservation_by_guest
func (sr *ReservationRepo) InsertReservationByGuest(guestReservation *ReservationByGuestCreate) error {
	reservationId := gocql.TimeUUID()

	err := sr.session.Query(
		`INSERT INTO reservations_by_guest 
         (reservation_id, guest_id,accommodation_id, accommodation_name,accommodation_location, check_in_date, check_out_date) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		reservationId,
		guestReservation.GuestId,
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

func (sr *ReservationRepo) GetReservationsByGuest(id string) (ReservationsByGuest, error) {
	scanner := sr.session.Query(`SELECT reservation_id, guest_id, accommodation_id, 
       accommodation_name, accommodation_location,
      check_in_date, check_out_date
FROM reservations_by_guest WHERE guest_id = ?`,
		id).Iter().Scanner()

	var reservations ReservationsByGuest
	for scanner.Next() {
		var rsv ReservationByGuest
		err := scanner.Scan(&rsv.ReservationId, &rsv.GuestId,
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

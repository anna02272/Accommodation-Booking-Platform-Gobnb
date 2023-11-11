package domain

import (
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
)

// NoSQL: AccommodationRepo struct encapsulating Cassandra api client
type AccommodationRepo struct {
	session *gocql.Session //connection towards CassandraDB
	logger  *log.Logger    //write messages inside Console
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
// if CassandrDB exists, this function connects to DB,if not it tries to create cassandraDB
func New(logger *log.Logger) (*AccommodationRepo, error) {
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
					}`, "accommodation", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to reservation keyspace
	cluster.Keyspace = "accommodation"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &AccommodationRepo{
		session: session,
		logger:  logger,
	}, nil
}

// Disconnect from database
func (sr *AccommodationRepo) CloseSession() {
	sr.session.Close()
}

// Create accommodation table
func (sr *AccommodationRepo) CreateTable() {
	err := sr.session.Query(
		`CREATE TABLE IF NOT EXISTS accommodation.accommodations
			(accommodationId UUID,
			accommodation_name text,
			accommodation_location text,
			accommodation_amenities text,
			accommodation_min_guests int,
			accommodation_max_guests int,
			accommodation_image_url text,
			PRIMARY KEY (accommodationId))`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
}

// inserting accommodation into table accommodation
func (sr *AccommodationRepo) InsertAccommodation(accommodation *Accommodation) error {
	accommodationId := gocql.TimeUUID()

	err := sr.session.Query(
		`INSERT INTO accommodations 
         (accommodationId, accommodation_name, accommodation_location, accommodation_amenities, accommodation_min_guests, accommodation_max_guests, accommodation_image_url) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		accommodationId,
		accommodation.Name,
		accommodation.Location,
		accommodation.Amenities,
		accommodation.MinGuests,
		accommodation.MaxGuests,
		accommodation.ImageUrl,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
		return err
	}

	return nil
}

func (sr *AccommodationRepo) GetAccommodations(id string) (Accommodations, error) {
	scanner := sr.session.Query(`SELECT accommodationId, 
        accommodation_name, accommodation_location, accommodation_amenities, accommodation_min_guests, accommodation_max_guests, accommodation_image_url
        FROM accommodation.accommodations WHERE accommodationId = ?`,
		id).Iter().Scanner()

	var accommodations Accommodations
	for scanner.Next() {
		var acm Accommodation
		err := scanner.Scan(&acm.AccommodationId, &acm.Name, &acm.Location)
		if err != nil {
			sr.logger.Println(err)
			return nil, err
		}
		accommodations = append(accommodations, &acm)
	}
	if err := scanner.Err(); err != nil {
		sr.logger.Println(err)
		return nil, err
	}
	return accommodations, nil
}

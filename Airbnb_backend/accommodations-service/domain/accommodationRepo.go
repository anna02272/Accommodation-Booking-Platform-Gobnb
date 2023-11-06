package domain

import (
	"fmt"
	"github.com/gocql/gocql"
	"log"
	"os"
)

// NoSQL: AccommodationsRepo struct encapsulating Cassandra api client
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
	// Create 'accommodation' keyspace
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

	// Connect to accommodation keyspace
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
		`CREATE TABLE IF NOT EXISTS accommodation (
        accommodationId UUID,
        accommodation_name text,
        accommodation_location text
        PRIMARY KEY (raccommodationId)
    ) WITH CLUSTERING ORDER BY (accommodationId ASC);`,
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
         (accommodationId, accommodation_name,accommodation_location) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
		accommodationId,
		accommodation.Name,
		accommodation.Location,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
		return err
	}

	return nil
}

func (sr *AccommodationRepo) GetAccommodations(id string) (Accommodations, error) {
	scanner := sr.session.Query(`SELECT accommodation_id, 
       accommodation_name, accommodation_location
FROM accommodations WHERE accommodation_id = ?`,
		id).Iter().Scanner()

	var accommodations Accommodations
	for scanner.Next() {
		var acm Accommodation
		err := scanner.Scan(&acm.AccommodationId, &acm.Name,
			&acm.Location)
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

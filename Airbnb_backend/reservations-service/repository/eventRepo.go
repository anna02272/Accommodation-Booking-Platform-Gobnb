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
)

type EventRepo struct {
	session *gocql.Session //connection towards CassandraDB
	logger  *log.Logger
	ctx     context.Context
	Tracer  trace.Tracer
}

// NoSQL: Constructor which reads db configuration from environment and creates a keyspace
// if CassandrDB exists, this function connects to DB,if not it tries to create cassandraDB
func NewEventRepo(logger *log.Logger, tracer trace.Tracer) (*EventRepo, error) {
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
					}`, "event", 1)).Exec()
	if err != nil {
		logger.Println(err)
	}
	session.Close()

	// Connect to reservation keyspace
	cluster.Keyspace = "event"
	cluster.Consistency = gocql.One
	session, err = cluster.CreateSession()
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	// Return repository with logger and DB session
	return &EventRepo{
		session: session,
		logger:  logger,
		Tracer:  tracer,
	}, nil
}

// Disconnect from database
func (sr *EventRepo) CloseSessionEvent() {
	sr.session.Close()
}

func (sr *EventRepo) CreateTableEventStore() {
	err := sr.session.Query(
		`CREATE TABLE IF NOT EXISTS event_store (
        event_id_time_created timeuuid,
        event text,
        guest_id text,
        accommodation_id text,
        PRIMARY KEY ((guest_id, event_id_time_created),accommodation_id)
    ) WITH CLUSTERING ORDER BY (accommodation_id ASC);`,
	).Exec()

	if err != nil {
		sr.logger.Println(err)
	}
	if err != nil {
		//span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
	}
}

func (sr *EventRepo) InsertEvent(ctx context.Context, eventData *data.AccommodationEvent) error {
	ctx, span := sr.Tracer.Start(ctx, "EventRepository.InsertEvent")
	defer span.End()

	eventID := gocql.TimeUUID()

	err := sr.session.Query(
		`INSERT INTO event_store 
         (event_id_time_created,event, guest_id,accommodation_id) 
         VALUES (?, ?, ?, ?)`,
		eventID,
		eventData.Event,
		eventData.GuestID,
		eventData.AccommodationID,
	).WithContext(ctx).Exec()

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		sr.logger.Println(err)
		return err
	}

	return nil
}

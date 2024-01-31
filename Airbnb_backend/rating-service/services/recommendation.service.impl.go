package services

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"rating-service/domain"
)

type RecommendationServiceImpl struct {
	driver neo4j.DriverWithContext
	//trace  trace.Tracer
	logger *log.Logger
}

func NewRecommendationServiceImpl(driver neo4j.DriverWithContext, trace trace.Tracer, logger *log.Logger) *RecommendationServiceImpl {
	uri := os.Getenv("NEO4J_DB")
	user := os.Getenv("NEO4J_USERNAME")
	pass := os.Getenv("NEO4J_PASS")
	auth := neo4j.BasicAuth(user, pass, "")
	log.Println("HEEEEEEEEEEEEEEEJJJJJJJJJJJJJJ")
	log.Println(auth)
	log.Println(uri)

	driver, err := neo4j.NewDriverWithContext(uri, auth)
	if err != nil {
		logger.Panic(err)
		return nil
	}

	// Return repository with logger and DB session
	return &RecommendationServiceImpl{
		driver: driver,
		logger: logger,
	}
}

// Check if connection is established
func (r *RecommendationServiceImpl) CheckConnection() {
	ctx := context.Background()
	err := r.driver.VerifyConnectivity(ctx)
	if err != nil {
		r.logger.Panic(err)
		return
	}
	// Print Neo4J server address
	r.logger.Printf(`Neo4J server address: %s`, r.driver.Target().Host)
}

// Disconnect from database
func (r *RecommendationServiceImpl) CloseDriverConnection(ctx context.Context) {
	r.driver.Close(ctx)
}
func (r *RecommendationServiceImpl) CreateUser(user *domain.NeoUser) error {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	savedMovie, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"CREATE (u:User) SET u.id = $id, u.username = $username, u.email = $email RETURN u.username + ', from node ' + id(u)",
				map[string]interface{}{"username": user.Username, "email": user.Email, "id": user.ID})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		r.logger.Println("Error inserting User:", err)
		return err
	}
	r.logger.Println(savedMovie.(string))
	return nil
}
func (r *RecommendationServiceImpl) CreateReservation(reservation *domain.ReservationByGuest) error {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	savedReservation, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (u:User {id: $guestId}), (a:Accommodation {accommodationId: $accommodationId}) "+
					"CREATE (r:Reservation) SET "+
					"r.reservationIdTimeCreated = timestamp(), "+
					"r.guestId = $guestId, "+
					"r.accommodationId = $accommodationId, "+
					"r.accommodationName = $accommodationName, "+
					"r.accommodationLocation = $accommodationLocation, "+
					"r.accommodationHostId = $accommodationHostId, "+
					"r.checkInDate = $checkInDate, "+
					"r.checkOutDate = $checkOutDate, "+
					"r.numberOfGuests = $numberOfGuests "+
					"CREATE (u)-[:REZERVISAO]->(r)-[:ZA_SMESTAJ]->(a) "+
					"CREATE (u)<-[:REZERVISAO]-(r)<-[:ZA_SMESTAJ]-(a) "+
					"RETURN r.reservationIdTimeCreated + ', from node ' + id(r)",

				map[string]interface{}{
					"guestId":               reservation.GuestId,
					"accommodationId":       reservation.AccommodationId,
					"accommodationName":     reservation.AccommodationName,
					"accommodationLocation": reservation.AccommodationLocation,
					"accommodationHostId":   reservation.AccommodationHostId,
					"checkInDate":           reservation.CheckInDate,
					"checkOutDate":          reservation.CheckOutDate,
					"numberOfGuests":        reservation.NumberOfGuests,
				})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		r.logger.Println("Error inserting Reservation:", err)
		return err
	}
	r.logger.Println(savedReservation.(string))
	return nil
}
func (r *RecommendationServiceImpl) CreateAccommodation(accommodation *domain.AccommodationRec) error {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	savedAccommodation, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"CREATE (a:Accommodation) SET a.accommodationId = $id,"+
					"a.hostId = $hostId,"+
					"a.name = $name,"+
					"a.location = $location,"+
					"a.minGuests = $minGuests,"+
					"a.maxGuests = $maxGuests,"+
					"a.active = $active"+
					" RETURN a.accommodationId + ', from node ' + id(a)",
				map[string]interface{}{
					"id":        accommodation.ID,
					"hostId":    accommodation.HostId,
					"name":      accommodation.Name,
					"location":  accommodation.Location,
					"minGuests": accommodation.MinGuests,
					"maxGuests": accommodation.MaxGuests,
					"active":    accommodation.Active,
				})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		r.logger.Println("Error inserting Accommodation:", err)
		return err
	}
	fmt.Printf("savedAccommodation", savedAccommodation)
	if savedAccommodation != nil {

		r.logger.Println(savedAccommodation.(string))
	} else {
		r.logger.Println("savedAccommodation is nil")
	}

	return nil
}

func (r *RecommendationServiceImpl) DeleteAccommodation(accommodationID string) error {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (a:Accommodation) WHERE a.accommodationId = $id DELETE a",
				map[string]interface{}{
					"id": accommodationID,
				})
			if err != nil {
				return nil, err
			}

			return nil, result.Err()
		})
	if err != nil {
		r.logger.Println("Error deleting Accommodation:", err)
		return err
	}

	return nil
}
func (r *RecommendationServiceImpl) CreateRate(rate *domain.RateAccommodationRec) error {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	savedAccommodation, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (u:User {id: $guestId}), (a:Accommodation {accommodationId: $accommodation}) "+
					"CREATE (r:Rate) SET r.rateId = $id,"+
					"r.guestId = $guestId,"+
					"r.accommodation = $accommodation,"+
					"r.rating = $rating "+
					"CREATE (u)-[:OCENJUJE]->(r)-[:OCENJEN]->(a) "+
					"CREATE (u)<-[:OCENJUJE]-(r)<-[:OCENJEN]-(a) "+
					" RETURN r.rateId + ', from node ' + id(r)",
				map[string]interface{}{
					"id":            rate.ID,
					"guestId":       rate.Guest,
					"accommodation": rate.Accommodation,
					"rating":        rate.Rating,
				})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		r.logger.Println("Error inserting Rate:", err)
		return err
	}
	r.logger.Println(savedAccommodation.(string))
	return nil
}
func (r *RecommendationServiceImpl) GetRecommendation(idd string) ([]domain.AccommodationRec, error) {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	log.Println("STIGAO U RECOMMENDATION")
	log.Println(idd)
	recommendation, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (trenutniKorisnik:User {id: $id})-[:REZERVISAO]->(rezervacija:Reservation)-[:ZA_SMESTAJ]->(smestaj:Accommodation)"+
					" MATCH (similarUser:User)-[:REZERVISAO]->(rezervacijaa:Reservation)-[:ZA_SMESTAJ]->(smestaj)"+
					" WHERE trenutniKorisnik <> similarUser"+
					" MATCH (similarUser) - [:REZERVISAO] ->(r:Reservation)-[:ZA_SMESTAJ]->(s:Accommodation)"+
					" WHERE smestaj <> s"+
					" MATCH (s)-[:OCENJEN]->(ocena:Rate)"+
					" WHERE ocena.rating>3"+
					" RETURN s.name as name, s.location as location, s.minGuests as minGuests, s.maxGuests as maxGuests,s.accommodationId as accommodationId,s.hostId as hostId,s.active as active",
				map[string]interface{}{
					"id": idd,
				})
			if err != nil {
				return nil, err
			}
			var news []domain.AccommodationRec

			for result.Next(ctx) {
				record := result.Record()
				name, _ := record.Get("name")
				id, _ := record.Get("accommodationId")
				location, _ := record.Get("location")
				hostId, _ := record.Get("hostId")
				minGuests, _ := record.Get("minGuests")
				maxGuests, _ := record.Get("maxGuests")
				active, _ := record.Get("active")

				new := domain.AccommodationRec{
					Name:      name.(string),
					ID:        id.(string),
					Location:  location.(string),
					HostId:    hostId.(string),
					MinGuests: int(minGuests.(int64)),
					MaxGuests: int(maxGuests.(int64)),
					Active:    active.(bool),
				}
				log.Println(new)
				news = append(news, new)
			}
			log.Println(news)
			return news, nil
		})
	if err != nil {
		r.logger.Println("Error get Recomm:", err)
		return []domain.AccommodationRec{}, nil
	}
	r.logger.Println(recommendation.([]domain.AccommodationRec))
	return recommendation.([]domain.AccommodationRec), nil
}

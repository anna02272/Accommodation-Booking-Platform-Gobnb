package services

import (
	"context"
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

//func New(logger *log.Logger) (*RecommendationServiceImpl, error) {
//	// Local instance
//	uri := os.Getenv("NEO4J_DB")
//	user := os.Getenv("NEO4J_USERNAME")
//	pass := os.Getenv("NEO4J_PASS")
//	auth := neo4j.BasicAuth(user, pass, "")
//	log.Println("HEEEEEEEEEEEEEEEJJJJJJJJJJJJJJ")
//	log.Println(auth)
//	log.Println(uri)
//
//	driver, err := neo4j.NewDriverWithContext(uri, auth)
//	if err != nil {
//		logger.Panic(err)
//		return nil, err
//	}
//
//	// Return repository with logger and DB session
//	return &RecommendationServiceImpl{
//		driver: driver,
//		logger: logger,
//	}, nil
//}

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
				"CREATE (u:User) SET u.username = $username, u.email = $email RETURN u.username + ', from node ' + id(u)",
				map[string]interface{}{"username": user.Username, "email": user.Email})
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
	//ctx := context.Background()
	//session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	//defer session.Close(ctx)
	//
	//// Begin a new transaction
	//transaction, err := session.BeginTransaction(neo4j.WriteAccess, neo4j.TxConfig{})
	//if err != nil {
	//	r.logger.Println("Error beginning transaction:", err)
	//	return err
	//}
	//defer transaction.Close()
	//
	//// Run the transaction logic
	//result, err := transaction.Run(ctx,
	//	"CREATE (u:User) SET u.username = $username, u.email = $email RETURN u.username + ', from node ' + id(u)",
	//	map[string]interface{}{"username": user.Username, "email": user.Email})
	//if err != nil {
	//	r.logger.Println("Error running transaction:", err)
	//	return err
	//}
	//
	//// Check for the next record in the result
	//if result.Next(ctx) {
	//	// Process the result if needed
	//	// Note: You may want to handle the result here
	//}
	//
	//// Commit the transaction
	//err = transaction.Commit()
	//if err != nil {
	//	r.logger.Println("Error committing transaction:", err)
	//	return err
	//}
	//
	//return nil
}

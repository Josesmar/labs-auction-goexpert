package auction

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id,omitempty"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      string                          `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionRepository struct {
	Collection *mongo.Collection
	mutex      sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
	go repo.StartAuctionClosureWorker()
	return repo
}

func getAuctionDuration() time.Duration {
	durationStr := os.Getenv("AUCTION_DURATION")
	if durationStr == "" {
		durationStr = "10m"
	}
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 10 * time.Minute
	}
	return duration
}

func (ar *AuctionRepository) StartAuctionClosureWorker() {
	interval := getAuctionInterval()
	for {
		time.Sleep(interval)
		ar.CloseExpiredAuctions()
	}
}

func getAuctionInterval() time.Duration {
	intervalStr := os.Getenv("AUCTION_INTERVAL")
	if intervalStr == "" {
		intervalStr = "20s"
	}
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return 20 * time.Second
	}
	return interval
}

func (ar *AuctionRepository) CloseExpiredAuctions() {
	ar.mutex.Lock()
	defer ar.mutex.Unlock()

	ctx := context.Background()
	cutoffTime := time.Now().Unix()

	filter := bson.M{
		"status":    "active",
		"timestamp": bson.M{"$lt": cutoffTime},
	}

	update := bson.M{
		"$set": bson.M{"status": "closed"},
	}

	result, err := ar.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		logger.Error("Error closing expired auctions", err)
		return
	}

	if result.ModifiedCount > 0 {
		logger.Info(fmt.Sprintf("Closed %d expired auctions", result.ModifiedCount))
	}
}

func NewMongoDBConnection(ctx context.Context) (*mongo.Database, error) {
	mongoURL := os.Getenv("MONGODB_URL")
	databaseName := os.Getenv("MONGODB_DB")
	if mongoURL == "" {
		return nil, fmt.Errorf("MONGODB_URL is not set")
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	if databaseName == "" {
		return nil, fmt.Errorf("MONGODB_DB is not set")
	}
	return client.Database(databaseName), nil
}

func (ar *AuctionRepository) CreateAuction(ctx context.Context, auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	statusMap := map[auction_entity.AuctionStatus]string{
		0: "active",
		1: "closed",
	}

	auctionEntityMongo := AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      statusMap[auctionEntity.Status],
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}

	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	return nil
}

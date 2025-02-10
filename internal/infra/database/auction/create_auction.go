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

	"go.mongodb.org/mongo-driver/mongo"
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
	Collection      *mongo.Collection
	auctionDuration time.Duration
	mutex           sync.Mutex
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(ctx context.Context, auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	statusMap := map[auction_entity.AuctionStatus]string{
		0: "active",
		1: "closed",
	}
	auctionEntityMongo := &AuctionEntityMongo{
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

	auctionEntity.Timestamp = time.Now().Add(ar.auctionDuration)
	go ar.scheduleAuctionClosure(auctionEntity.Id, auctionEntity.Timestamp)

	return nil
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

func (ar *AuctionRepository) scheduleAuctionClosure(auctionID string, endTime time.Time) {
	duration := time.Until(endTime)
	time.Sleep(duration)
	ar.mutex.Lock()
	defer ar.mutex.Unlock()
	ctx := context.Background()
	_, err := ar.Collection.UpdateOne(ctx, map[string]string{"_id": auctionID}, map[string]interface{}{"$set": map[string]string{"status": "closed"}})
	if err != nil {
		logger.Error("Error trying to close auction", err)
		return
	}
	logger.Info(fmt.Sprintf("Auction %s has ended", auctionID))
}

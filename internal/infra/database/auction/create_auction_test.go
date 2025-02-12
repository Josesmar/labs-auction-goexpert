package auction

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCloseExpiredAuctions(t *testing.T) {
	ctx := context.Background()
	db, err := NewMongoDBConnection(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	repo := NewAuctionRepository(db)

	repo.Collection.DeleteOne(ctx, bson.M{"_id": "test-auction"})

	_, err = repo.Collection.InsertOne(ctx, bson.M{
		"_id":       "test-auction",
		"status":    "active",
		"timestamp": time.Now().Add(-15 * time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("Failed to insert test auction: %v", err)
	}

	repo.CloseExpiredAuctions()

	var auction bson.M
	err = repo.Collection.FindOne(ctx, bson.M{"_id": "test-auction"}).Decode(&auction)
	if err != nil {
		t.Fatalf("Failed to find test auction: %v", err)
	}

	if auction["status"] != "closed" {
		t.Errorf("Expected auction to be closed, but got %v", auction["status"])
	}
}

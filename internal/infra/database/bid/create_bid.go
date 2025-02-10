package bid

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/entity/bid_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type BidEntityMongo struct {
	Id        string  `bson:"_id"`
	UserId    string  `bson:"user_id"`
	AuctionId string  `bson:"auction_id"`
	Amount    float64 `bson:"amount"`
	Timestamp int64   `bson:"timestamp"`
}

type BidRepository struct {
	Collection            *mongo.Collection
	AuctionRepository     *auction.AuctionRepository
	auctionInterval       time.Duration
	auctionStatusMap      map[string]auction_entity.AuctionStatus
	auctionEndTimeMap     map[string]time.Time
	auctionStatusMapMutex *sync.Mutex
	auctionEndTimeMutex   *sync.Mutex
}

func NewBidRepository(database *mongo.Database, auctionRepository *auction.AuctionRepository) *BidRepository {
	return &BidRepository{
		auctionInterval:       getAuctionInterval(),
		auctionStatusMap:      make(map[string]auction_entity.AuctionStatus),
		auctionEndTimeMap:     make(map[string]time.Time),
		auctionStatusMapMutex: &sync.Mutex{},
		auctionEndTimeMutex:   &sync.Mutex{},
		Collection:            database.Collection("bids"),
		AuctionRepository:     auctionRepository,
	}
}

func (bd *BidRepository) CreateBid(ctx context.Context, bidEntities []bid_entity.Bid) *internal_error.InternalError {
	for _, bid := range bidEntities {
		bd.auctionStatusMapMutex.Lock()
		auctionStatus, okStatus := bd.auctionStatusMap[bid.AuctionId]
		bd.auctionStatusMapMutex.Unlock()

		bd.auctionEndTimeMutex.Lock()
		auctionEndTime, okEndTime := bd.auctionEndTimeMap[bid.AuctionId]
		bd.auctionEndTimeMutex.Unlock()

		if okEndTime && okStatus {
			if auctionStatus == auction_entity.Completed || time.Now().After(auctionEndTime) {
				logger.Info(fmt.Sprintf("Auction %s completed", bid.AuctionId))
				continue
			}
		} else {
			auctionEntity, err := bd.AuctionRepository.FindAuctionById(ctx, bid.AuctionId)
			if err != nil {
				logger.Error(fmt.Sprintf("Erro ao buscar leilão ID %s", bid.AuctionId), err)
				return internal_error.NewInternalServerError("Erro ao buscar leilão")
			}
			if auctionEntity.Status == auction_entity.Completed {
				logger.Info(fmt.Sprintf("Auction %s is completed, bid rejected", bid.AuctionId))
				continue
			}

			bd.auctionStatusMapMutex.Lock()
			bd.auctionStatusMap[bid.AuctionId] = auctionEntity.Status
			bd.auctionStatusMapMutex.Unlock()

			bd.auctionEndTimeMutex.Lock()
			bd.auctionEndTimeMap[bid.AuctionId] = auctionEntity.Timestamp
			bd.auctionEndTimeMutex.Unlock()
		}

		bidEntityMongo := &BidEntityMongo{
			Id:        bid.Id,
			UserId:    bid.UserId,
			AuctionId: bid.AuctionId,
			Amount:    bid.Amount,
			Timestamp: bid.Timestamp.Unix(),
		}

		_, err := bd.Collection.InsertOne(ctx, bidEntityMongo)
		if err != nil {
			logger.Error("Error inserting bid entity", err)
			return internal_error.NewInternalServerError("Erro ao inserir bid no banco de dados")
		}

		logger.Info("✅ Bid inserido com sucesso no MongoDB")
	}

	return nil
}

func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5 // Valor padrão caso a conversão falhe
	}
	return duration
}

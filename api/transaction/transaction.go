package transaction

import (
	"Fortune_Tracker_API/api/ledger"
	"Fortune_Tracker_API/pkg/logger"
	"Fortune_Tracker_API/pkg/mongodb"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type transactionType struct {
	Action     string `json:"Action" bson:"Action" binding:"required"`
	ParentType uint8  `json:"ParentType" bson:"ParentType" binding:"required"`
	ChildType  uint8  `json:"ChildType" bson:"ChildType" binding:"required"`
}

type transactionSharer struct {
	UUID   string  `json:"UUID" bson:"UUID" binding:"required"`
	Amount float64 `json:"Amount" bson:"Amount" binding:"required"`
}

type transaction struct {
	UTID       string              `json:"UTID" bson:"UTID"`
	ULID       string              `json:"ULID" bson:"ULID"`
	Amount     float64             `json:"Amount" bson:"Amount" binding:"required"`
	RecordTime uint32              `json:"RecordTime" bson:"RecordTime" binding:"required"`
	UpdateTime uint32              `json:"UpdateTime" bson:"UpdateTime" binding:"required"`
	Type       transactionType     `json:"Type" bson:"Type"`
	Name       string              `json:"Name" bson:"Name" binding:"required"`
	Payer      string              `json:"Payer" bson:"Payer" binding:"required"`
	Sharers    []transactionSharer `json:"Sharers" bson:"Sharers" binding:"required"`
}

func create(ts transaction, UUID string) (string, error) {
	// Get members of ledger
	var err error
	var members map[string]bool
	if members, err = ledger.GetLedgerMember(ts.ULID); err != nil {
		return "", err
	} 

	// These users should be the member of the ledger
	// -> UUID in token, payer, all user in sharers
	if !members[UUID] {
		logger.Warn("[TRANSACTION] User is not a member of the ledger")
		return "", errors.New("user is not a member of the ledger")
	} else if !members[ts.Payer] {
		logger.Warn("[TRANSACTION] Payer is not a member of the ledger")
		return "", errors.New("payer is not a member of the ledger")
	} 

	for _, sharer := range ts.Sharers {
		if !members[sharer.UUID] {
			logger.Warn("[TRANSACTION] A sharer is not a member of the ledger")
			return "", errors.New("a sharer is not a member of the ledger")
		}
	}

	// Generate UTID
	ts.UTID = uuid.New().String()

	// Insert transaction into mongodb transaction collection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = mongodb.TransactionCollection.InsertOne(ctx, ts)
	if err != nil {
		logger.Error("[TRANSACTION] " + err.Error())
		return "", err
	}

	logger.Info("[TRANSACTION] Transaction:" + ts.UTID + " created")

	return ts.UTID, nil
}

func deleteT(ULID, UTID, UUID string) error {
	// Check the user is in the ledger of the transaction
	var err error
	var members map[string]bool
	if members, err = ledger.GetLedgerMember(ULID); err != nil {
		return err
	} 

	if !members[UUID] {
		logger.Warn("[TRANSACTION] User is not a member of the ledger")
		return errors.New("user is not a member of the ledger")
	}

	// Delete the transaction from mongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"UTID": UTID}

	result, err := mongodb.TransactionCollection.DeleteOne(ctx, filter)
	if err != nil {
		logger.Error("[TRANSACTION] " + err.Error())
		return err
	} else if result.DeletedCount == 0 {
		logger.Warn("[TRANSACTION] Transaction not found")
		return errors.New("transaction not found")
	}

	logger.Info("[TRANSACTION] Transaction:" + UTID + " deleted")

	return nil
}

func get(ULID, UTID, UUID string) (transaction, error) {
	// Check the user is in the ledger of the transaction
	var err error
	var ts transaction
	var members map[string]bool
	if members, err = ledger.GetLedgerMember(ULID); err != nil {
		return ts, err
	} 

	if !members[UUID] {
		logger.Warn("[TRANSACTION] User is not a member of the ledger")
		return ts, errors.New("user is not a member of the ledger")
	}

	// Get the transaction from mongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"UTID": UTID}

	err = mongodb.TransactionCollection.FindOne(ctx, filter).Decode(&ts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("[TRANSACTION] Transaction not found")
			return ts, errors.New("transaction not found")
		}
		logger.Error("[TRANSACTION] " + err.Error())
		return transaction{}, err
	}

	logger.Info("[TRANSACTION] Transaction:" + UTID + " retrieved")

	return ts, nil
}

// get the transaction between startTime and endTime in the given ledger
func getByTime(gbtr getByTimeRequest, UUID string) ([]transaction, error) {
	// Check the user is in the ledger of the transaction
	var err error
	var tss []transaction
	var members map[string]bool
	if members, err = ledger.GetLedgerMember(gbtr.ULID); err != nil {
		return tss, err
	} 

	if !members[UUID] {
		logger.Warn("[TRANSACTION] User is not a member of the ledger")
		return tss, errors.New("user is not a member of the ledger")
	}

	// Get the transaction from mongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"ULID": gbtr.ULID,
		"RecordTime": bson.M{
			"$gte": gbtr.StartTime,
			"$lte": gbtr.EndTime,
		},
	}

	cursor, err := mongodb.TransactionCollection.Find(ctx, filter)
	if err != nil {
		logger.Error("[TRANSACTION] " + err.Error())
		return tss, err
	}

	if err = cursor.All(ctx, &tss); err != nil {
		logger.Error("[TRANSACTION] " + err.Error())
		return tss, err
	}

	logger.Info("[TRANSACTION] Transactions retrieved")

	return tss, nil
}

func update(UUID string, ts transaction) error {
	// Check the user is in the ledger of the transaction
	var err error
	var members map[string]bool
	if members, err = ledger.GetLedgerMember(ts.ULID); err != nil {
		return err
	} 

	if !members[UUID] {
		logger.Warn("[TRANSACTION] User is not a member of the ledger")
		return errors.New("user is not a member of the ledger")
	}

	// Get the transaction from mongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"UTID": ts.UTID,
	}

	update := bson.M{
		"$set": bson.M{
			"Amount":     ts.Amount,
			"RecordTime": ts.RecordTime,
			"UpdateTime": ts.UpdateTime,
			"Type":       ts.Type,
			"Name":       ts.Name,
			"Payer":      ts.Payer,
			"Sharers":    ts.Sharers,
		},
	}

	result, err := mongodb.TransactionCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("[TRANSACTION] " + err.Error())
		return err
	} else if result.MatchedCount == 0 {
		logger.Warn("[TRANSACTION] Transaction not found")
		return errors.New("transaction not found")
	}

	logger.Info("[TRANSACTION] Transaction:" + ts.UTID + " updated")

	return nil
}

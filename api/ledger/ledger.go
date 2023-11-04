package ledger

import (
	"Fortune_Tracker_API/pkg/logger"
	"Fortune_Tracker_API/pkg/mongodb"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type member struct {
	UUID     string `json:"UUID" bson:"UUID"`
	Nickname string `json:"Nickname" bson:"Nickname"`
}

type childType struct {
	CTID int    `json:"CTID" bson:"CTID"`
	Name string `json:"Name" bson:"Name"`
}

type parentType struct {
	PTID       int         `json:"PTID" bson:"PTID"`
	Name       string      `json:"Name" bson:"Name"`
	ChildTypes []childType `json:"ChildTypes" bson:"ChildTypes"`
}

type ledgerType struct {
	ParentTypes []parentType `json:"ParentTypes" bson:"ParentTypes"`
}

type ledger struct {
	ULID         string     `json:"ULID" bson:"ULID"`
	Name         string     `json:"Name" bson:"Name" binding:"required"`
	Notification bool       `json:"Notification" bson:"Notification" binding:"required"`
	Theme        string     `json:"Theme" bson:"Theme" binding:"required"`
	Currency     string     `json:"Currency" bson:"Currency" binding:"required"`
	Types        ledgerType `json:"Types" bson:"Types" binding:"required"`
	Members      []member   `json:"Members" bson:"Members" binding:"required"`
}

func GetLedgerMember(ULID string) (map[string]bool, error) {
	// Get ledger members
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"ULID": ULID,
	}

	var ledger ledger
	err := mongodb.LedgerCollection.FindOne(ctx, filter).Decode(&ledger)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("[LEDGER] Ledger not found")
			return nil, errors.New("ledger not found")
		}
		logger.Error("[LEDGER] " + err.Error())
		return nil, err
	}

	// only return the UUIDs map
	uuids := make(map[string]bool)
	for _, member := range ledger.Members {
		uuids[member.UUID] = true
	}

	return uuids, nil
}

func CheckLedgerExists(ULID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// check ledger exists first
	filter := bson.M{
		"ULID": ULID,
	}

	var ledger ledger
	err := mongodb.LedgerCollection.FindOne(ctx, filter).Decode(&ledger)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger.Warn("[LEDGER] Ledger not found")
			return false, nil
		}
		logger.Error("[LEDGER] " + err.Error())
		return false, err
	}

	return true, nil
}

func create(l ledger) (string, error) {
	// Genrate ULID
	l.ULID = uuid.NewString()

	// Insert into ledger database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := mongodb.LedgerCollection.InsertOne(ctx, &l)
	if err != nil {
		logger.Error("[LEDGER] " + err.Error())
		return "", err
	}

	logger.Info("[LEDGER] Created ledger: " + l.ULID)

	return l.ULID, nil
}

func get(UUID string) ([]ledger, error) {
	var userLedgers []ledger

	// Get ledger info for a user
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"Members": bson.M{
			"$elemMatch": bson.M{
				"UUID": UUID,
			},
		},
	}

	cur, err := mongodb.LedgerCollection.Find(ctx, filter)
	if err != nil {
		logger.Error("[LEDGER] " + err.Error())
		return userLedgers, err
	}

	if err := cur.All(ctx, &userLedgers); err != nil {
		logger.Error("[LEDGER] " + err.Error())
		return userLedgers, err
	}

	logger.Info("[LEDGER] Get ledger info for user: " + UUID)

	return userLedgers, nil
}

func update(ur updateRequest, ULID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// define filter and update query
	filter := bson.M{
		"ULID": ULID,
	}

	update := bson.M{
		"$set": bson.M{},
	}

	// update fields if they are not empty
	if ur.Name != nil {
		update["$set"].(bson.M)["Name"] = *ur.Name
	}
	if ur.Notification != nil {
		update["$set"].(bson.M)["Notification"] = *ur.Notification
	}
	if ur.Theme != nil {
		update["$set"].(bson.M)["Theme"] = *ur.Theme
	}
	if ur.Currency != nil {
		update["$set"].(bson.M)["Currency"] = *ur.Currency
	}

	result := mongodb.LedgerCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate())
	if result.Err() != nil {
		logger.Error("[LEDGER] " + result.Err().Error())
		return result.Err()
	} else if result == nil {
		logger.Error("[LEDGER] " + "ledger not found")
		return result.Err()
	}

	logger.Info("[LEDGER] Updated ledger: " + ULID)

	return nil
}

func addMember(amr addMemberRequest, ULID string) error {
	// check ledger exists first
	if exist, err := CheckLedgerExists(ULID); err != nil {
		return err
	} else if !exist {
		return errors.New("ledger not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// add member to ledger if they are not already in it
	filter := bson.M{
		"ULID": ULID,
		"Members": bson.M{
			"$not": bson.M{
				"$elemMatch": bson.M{
					"UUID": amr.UUID,
				},
			},
		},
	}

	update := bson.M{
		"$push": bson.M{
			"Members": bson.M{
				"UUID":     amr.UUID,
				"Nickname": amr.Nickname,
			},
		},
	}

	result := mongodb.LedgerCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate())
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			logger.Error("[LEDGER] User already exists in the ledger")
			return errors.New("user already exists in the ledger")
		}
		logger.Error("[LEDGER] " + result.Err().Error())
		return result.Err()
	}

	logger.Info("[LEDGER] Added member to ledger: " + ULID)
	return nil
}

func removeMember(ULID, UUID string) error {
	// check ledger exists first
	if exist, err := CheckLedgerExists(ULID); err != nil {
		return err
	} else if !exist {
		return errors.New("ledger not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// delete the user from the ledger if they exist
	filter := bson.M{
		"ULID": ULID,
		"Members": bson.M{
			"$elemMatch": bson.M{
				"UUID": UUID,
			},
		},
	}

	update := bson.M{
		"$pull": bson.M{
			"Members": bson.M{
				"UUID": UUID,
			},
		},
	}

	result := mongodb.LedgerCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate())
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			logger.Error("[LEDGER] User not found in the ledger")
			return errors.New("user not found in the ledger")
		}
		logger.Error("[LEDGER] " + result.Err().Error())
		return result.Err()
	}

	logger.Info("[LEDGER] Removed member from ledger: " + ULID)
	return nil
}

func updateNickname(unr updateNicknameRequest, ULID, UUID string) error {
	// check ledger exists first
	if exist, err := CheckLedgerExists(ULID); err != nil {
		return err
	} else if !exist {
		return errors.New("ledger not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// update the user's nickname
	filter := bson.M{
		"ULID": ULID,
		"Members": bson.M{
			"$elemMatch": bson.M{
				"UUID": UUID,
			},
		},
	}

	update := bson.M{
		"$set": bson.M{
			"Members.$.Nickname": unr.Nickname,
		},
	}

	result := mongodb.LedgerCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate())
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			logger.Error("[LEDGER] User not found in the ledger")
			return errors.New("user not found in the ledger")
		}
		logger.Error("[LEDGER] " + result.Err().Error())
		return result.Err()
	}

	logger.Info("[LEDGER] Updated member nickname in ledger: " + ULID)
	return nil
}

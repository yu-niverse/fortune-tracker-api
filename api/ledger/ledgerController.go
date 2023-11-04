package ledger

import (
	"Fortune_Tracker_API/internal/response"
	"Fortune_Tracker_API/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type updateRequest struct {
	Name         *string `json:"Name" bson:"Name"`
	Notification *bool   `json:"Notification" bson:"Notification"`
	Theme        *string `json:"Theme" bson:"Theme"`
	Currency     *string `json:"Currency" bson:"Currency"`
}

type addMemberRequest struct {
	UUID     string `json:"UUID" bson:"UUID" binding:"required"`
	Nickname string `json:"Nickname" bson:"Nickname" binding:"required"`
}

type updateNicknameRequest struct {
	Nickname string `json:"Nickname" bson:"Nickname" binding:"required"`
}

func Create(c *gin.Context) {
	var err error
	var ULID string

	// Create response
	r := response.New()

	// Parse request body to JSON format
	var ledger ledger
	if err = c.ShouldBindJSON(&ledger); err != nil {
		logger.Warn("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Create ledger
	if ULID, err = create(ledger); err != nil {
		logger.Warn("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Return response
	r.Status = true
	r.Data = response.ULIDResponse{ULID: ULID}
	c.JSON(http.StatusOK, r)
}

func Get(c *gin.Context) {
	var err error
	var userLedgers []ledger

	// Create response
	r := response.New()

	// Get UUID 
	UUID := c.MustGet("UUID").(string)

	// Get ledger info
	if userLedgers, err = get(UUID); err != nil {
		logger.Warn("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Return response
	r.Status = true
	r.Data = userLedgers
	c.JSON(http.StatusOK, r)
}

func Update(c *gin.Context) {
	var err error
	var updateRequest updateRequest

	// Create response
	r := response.New()

	// Parse request body to JSON format
	if err = c.ShouldBindJSON(&updateRequest); err != nil {
		logger.Warn("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Update ledger
	if err = update(updateRequest, c.Param("ulid")); err != nil {
		logger.Error("[LEDGER] " + err.Error())
		if err == mongo.ErrNoDocuments {
			r.Message = "Ledger not found"
			c.JSON(http.StatusNotFound, r)
			return
		}
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	c.JSON(http.StatusOK, r)
}

func AddMember(c *gin.Context) {
	var err error

	// Create response
	r := response.New()

	// Parse request body to JSON format
	var addMemberRequest addMemberRequest
	if err = c.ShouldBindJSON(&addMemberRequest); err != nil {
		logger.Warn("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Add member
	if err = addMember(addMemberRequest, c.Param("ulid")); err != nil {
		if err.Error() == "ledger not found" {
			r.Message = "Ledger not found"
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user already exists in the ledger" {
			r.Message = err.Error()
			c.JSON(http.StatusBadRequest, r)
			return
		}
		logger.Error("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	c.JSON(http.StatusOK, r)
}

func RemoveMember(c *gin.Context) {
	var err error

	// Create response
	r := response.New()

	// Remove member
	if err = removeMember(c.Param("ulid"), c.MustGet("UUID").(string)); err != nil {
		if err.Error() == "ledger not found" {
			r.Message = "Ledger not found"
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user not found in the ledger" {
			r.Message = err.Error()
			c.JSON(http.StatusBadRequest, r)
			return
		}
		logger.Error("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	c.JSON(http.StatusNoContent, r)
}

func UpdateNickname(c *gin.Context) {
	var err error

	// Create response
	r := response.New()

	// Parse request body to JSON format
	var updateNicknameRequest updateNicknameRequest
	if err = c.ShouldBindJSON(&updateNicknameRequest); err != nil {
		logger.Warn("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Update nickname
	if err = updateNickname(updateNicknameRequest, c.Param("ulid"), c.MustGet("UUID").(string)); err != nil {
		if err.Error() == "ledger not found" {
			r.Message = "Ledger not found"
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user not found in the ledger" {
			r.Message = err.Error()
			c.JSON(http.StatusBadRequest, r)
			return
		}
		logger.Error("[LEDGER] " + err.Error())
		r.Message = err.Error()
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	c.JSON(http.StatusOK, r)
}

package transaction

import (
	"Fortune_Tracker_API/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type getByTimeRequest struct {
	ULID      string `json:"ULID" bson:"ULID"`
	StartTime uint32 `json:"StartTime" bson:"StartTime" binding:"required"`
	EndTime   uint32 `json:"EndTime" bson:"EndTime" binding:"required"`
}

func Create(c *gin.Context) {
	var err error
	var UTID string
	// Create response
	r := response.New()

	// Parse request body to JSON format
	var transaction transaction
	if err = c.ShouldBindJSON(&transaction); err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	transaction.ULID = c.Param("ulid")

	// Amount should be positive
	if transaction.Amount <= 0 {
		r.Message = "amount should be positive"
		c.JSON(http.StatusBadRequest, r)
		return
	}
	totalAmount := 0.0
	for _, sharer := range transaction.Sharers {
		totalAmount += sharer.Amount
		if sharer.Amount <= 0 {
			r.Message = "amount should be positive"
			c.JSON(http.StatusBadRequest, r)
			return
		}
	}

	// Amount should be equal to the sum of sharers' amount
	if totalAmount != transaction.Amount {
		r.Message = "amount should be equal to the sum of sharers' amount"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Type.Action should be "income" or "expense" or "transfer"
	if transaction.Type.Action != "income" &&
		transaction.Type.Action != "expense" &&
		transaction.Type.Action != "transfer" {
		r.Message = "type.action should be income or expense or transfer"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Record and update time should be in the past
	if transaction.RecordTime > uint32((uint64(time.Now().Unix())<<32>>32)) ||
		transaction.UpdateTime > uint32((uint64(time.Now().Unix())<<32>>32)) {
		r.Message = "record and update time should be in the past"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Create transaction
	if UTID, err = create(transaction, c.MustGet("UUID").(string)); err != nil {
		r.Message = err.Error()
		if err.Error() == "ledger not found" {
			c.JSON(http.StatusNotFound, r)
			return
		} else if strings.Contains(err.Error(), "is not a member of the ledger") {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	r.Data = response.UTIDResponse{UTID: UTID}
	c.JSON(http.StatusCreated, r)
}

func Delete(c *gin.Context) {
	var err error

	// Create response
	r := response.New()

	// Delete transaction
	if err = deleteT(c.Param("ulid"), c.Param("utid"), c.MustGet("UUID").(string)); err != nil {
		r.Message = err.Error()
		if err.Error() == "transaction not found" || err.Error() == "ledger not found" {
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user is not a member of the ledger" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	c.JSON(http.StatusNoContent, r)
}

func Get(c *gin.Context) {
	var err error
	var ts transaction

	// Create response
	r := response.New()

	// Get transactions
	if ts, err = get(c.Param("ulid"), c.Param("utid"), c.MustGet("UUID").(string)); err != nil {
		r.Message = err.Error()
		if err.Error() == "transaction not found" || err.Error() == "ledger not found" {
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user is not a member of the ledger" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	r.Data = ts
	c.JSON(http.StatusOK, r)
}

func GetByTime(c *gin.Context) {
	var err error
	var tss []transaction

	// Create response
	r := response.New()

	// Parse request body to JSON format
	var gbtr getByTimeRequest
	if err = c.ShouldBindJSON(&gbtr); err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	gbtr.ULID = c.Param("ulid")

	// Get transactions
	if tss, err = getByTime(gbtr, c.MustGet("UUID").(string)); err != nil {
		r.Message = err.Error()
		if err.Error() == "ledger not found" {
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user is not a member of the ledger" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	r.Data = tss
	c.JSON(http.StatusOK, r)
}

func Update(c *gin.Context) {
	var err error

	// Create response
	r := response.New()

	// Parse request body to JSON format
	var ts transaction
	if err = c.ShouldBindJSON(&ts); err != nil {
		r.Message = err.Error()
		c.JSON(http.StatusBadRequest, r)
		return
	}

	ts.ULID = c.Param("ulid")
	ts.UTID = c.Param("utid")

	// Amount should be positive
	if ts.Amount <= 0 {
		r.Message = "amount should be positive"
		c.JSON(http.StatusBadRequest, r)
		return
	}
	totalAmount := 0.0
	for _, sharer := range ts.Sharers {
		totalAmount += sharer.Amount
		if sharer.Amount <= 0 {
			r.Message = "amount should be positive"
			c.JSON(http.StatusBadRequest, r)
			return
		}
	}

	// Amount should be equal to the sum of sharers' amount
	if totalAmount != ts.Amount {
		r.Message = "amount should be equal to the sum of sharers' amount"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Type.Action should be "income" or "expense" or "transfer"
	if ts.Type.Action != "income" && ts.Type.Action != "expense" && ts.Type.Action != "transfer" {
		r.Message = "type.action should be income or expense or transfer"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Record and update time should be in the past
	if ts.RecordTime > uint32((uint64(time.Now().Unix())<<32>>32)) ||
		ts.UpdateTime > uint32((uint64(time.Now().Unix())<<32>>32)) {
		r.Message = "record and update time should be in the past"
		c.JSON(http.StatusBadRequest, r)
		return
	}

	// Update transactions
	if err = update(c.MustGet("UUID").(string), ts); err != nil {
		r.Message = err.Error()
		if err.Error() == "transaction not found" || err.Error() == "ledger not found" {
			c.JSON(http.StatusNotFound, r)
			return
		} else if err.Error() == "user is not a member of the ledger" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusInternalServerError, r)
		return
	}

	// Return response
	r.Status = true
	c.JSON(http.StatusOK, r)
}

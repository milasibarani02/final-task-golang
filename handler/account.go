package handler

import (
	"task-golang-db/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	
)

type AccountInterface interface {
	Create(*gin.Context)
	Read(*gin.Context)
	Update(*gin.Context)
	Delete(*gin.Context)
	List(*gin.Context)
	Topup(c *gin.Context)
	Transfer(c *gin.Context)
	Balance(c *gin.Context)
	My(*gin.Context)
	Mutation(*gin.Context)
}

type accountImplement struct {
	db *gorm.DB
}

func NewAccount(db *gorm.DB) AccountInterface {
	return &accountImplement{
		db: db,
	}
}

func (a *accountImplement) Create(c *gin.Context) {
	payload := model.Account{}

	// bind JSON Request to payload
	err := c.BindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	// Create data
	result := a.db.Create(&payload)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Create success",
		"data":    payload,
	})
}

func (a *accountImplement) Read(c *gin.Context) {
	var account model.Account

	// get id from url account/read/5, 5 will be the id
	id := c.Param("id")

	// Find first data based on id and put to account model
	if err := a.db.First(&account, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})
			return
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"data": account,
	})
}

func (a *accountImplement) Update(c *gin.Context) {
	payload := model.Account{}

	// bind JSON Request to payload
	err := c.BindJSON(&payload)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
		return
	}

	// get id from url account/update/5, 5 will be the id
	id := c.Param("id")

	// Find first data based on id and put to account model
	account := model.Account{}
	result := a.db.First(&account, "account_id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	// Update data
	account.Name = payload.Name
	a.db.Save(account)

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Update success",
	})
}

func (a *accountImplement) Delete(c *gin.Context) {
	// get id from url account/delete/5, 5 will be the id
	id := c.Param("id")

	// Find first data based on id and delete it
	if err := a.db.Where("account_id = ?", id).Delete(&model.Account{}).Error; err != nil {
		// No data found and deleted
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})
			return
		}
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete success",
		"data": map[string]string{
			"account_id": id,
		},
	})
}

func (a *accountImplement) List(c *gin.Context) {
	// Prepare empty result
	var accounts []model.Account

	// Find and get all accounts data and put to &accounts
	if err := a.db.Find(&accounts).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"data": accounts,
	})
}

func (a *accountImplement) My(c *gin.Context) {
	var account model.Account
	// get account_id from middleware auth
	accountID := c.GetInt64("account_id")

	// Find first data based on account_id given
	if err := a.db.First(&account, accountID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"data": account,
	})
}

func (a *accountImplement) Topup(c *gin.Context) {
	accountID := c.GetInt64("account_id")
	var payload struct {
		Amount int64 `json:"amount"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Amount <= 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid topup amount"})
		return
	}

	tx := a.db.Begin()

	// Update account balance
	if err := a.db.Model(&model.Account{}).Where("account_id = ?", accountID).
		Update("balance", gorm.Expr("balance + ?", payload.Amount)).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transaction := model.Transaction{
		AccountID:       accountID,
		Amount:          payload.Amount,
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Commit transaksi
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Topup successful"})
}

func (a *accountImplement) Transfer(c *gin.Context) {
	accountID := c.GetInt64("account_id")
	var payload struct {
		TargetAccountID int64 `json:"target_account_id"`
		Amount          int64 `json:"amount"`
	}

	if err := c.BindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Amount <= 0 || payload.TargetAccountID == accountID {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid transfer request"})
		return
	}

	// Check balance
	var account model.Account
	if err := a.db.First(&account, accountID).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if account.Balance < payload.Amount {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Update balances
	tx := a.db.Begin()
	if err := tx.Model(&model.Account{}).Where("account_id = ?", accountID).
		Update("balance", gorm.Expr("balance - ?", payload.Amount)).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Model(&model.Account{}).Where("account_id = ?", payload.TargetAccountID).
		Update("balance", gorm.Expr("balance + ?", payload.Amount)).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Catat transaksi pengirim
	transactionSender := model.Transaction{
		AccountID:       accountID,
		TransactionCategoryID: nil, // Sesuaikan dengan kategori transaksi
		Amount:          -payload.Amount, // Saldo berkurang untuk pengirim
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transactionSender).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Catat transaksi penerima
	transactionReceiver := model.Transaction{
		AccountID:       payload.TargetAccountID,
		TransactionCategoryID: nil, // Sesuaikan dengan kategori transaksi
		Amount:          payload.Amount, // Saldo bertambah untuk penerima
		TransactionDate: time.Now(),
	}
	if err := tx.Create(&transactionReceiver).Error; err != nil {
		tx.Rollback()
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
}

func (a *accountImplement) Balance(c *gin.Context) {
	accountID := c.GetInt64("account_id")

	var balance float64
	if err := a.db.Model(&model.Account{}).
		Where("account_id = ?", accountID).
		Select("balance").
		Scan(&balance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func (a *accountImplement) Mutation(c *gin.Context) {
	// Ambil account_id dari context setelah authentication
	accountID := c.GetInt64("account_id")

	var transactions []model.Transaction
	if err := a.db.Where("account_id = ?", accountID).
		Order("transaction_date DESC"). // Mengurutkan berdasarkan tanggal transaksi terbaru
		Limit(10). // Membatasi hasil ke 10 transaksi terakhir
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": transactions})
}

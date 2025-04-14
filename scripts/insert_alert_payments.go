package scripts

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InsertAlertPayments(dsn string) {
	// Connect to the database
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Get current timestamp
	now := time.Now().Unix()

	// Create payments for different gateways with high failure rates
	payments := []struct {
		ID         string `gorm:"column:id;primaryKey"`
		CreatedAt  int64  `gorm:"column:created_at"`
		PaymentID  string `gorm:"column:payment_id"`
		MerchantID string `gorm:"column:merchant_id"`
		Amount     int64  `gorm:"column:amount"`
		Currency   string `gorm:"column:currency"`
		Method     string `gorm:"column:method"`
		Status     string `gorm:"column:status"`
		Gateway    string `gorm:"column:gateway"`
		CapturedAt int64  `gorm:"column:captured_at"`
	}{
		// HDFC payments (high failure rate)
		{
			ID:         "pay_HDFC1",
			CreatedAt:  now - 1800, // 30 minutes ago
			PaymentID:  "pay_HDFC1",
			MerchantID: "merchant1",
			Amount:     10000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         "pay_HDFC2",
			CreatedAt:  now - 1200, // 20 minutes ago
			PaymentID:  "pay_HDFC2",
			MerchantID: "merchant1",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HDFC3",
			CreatedAt:  now - 600, // 10 minutes ago
			PaymentID:  "pay_HDFC3",
			MerchantID: "merchant1",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},

		// ICICI payments (high failure rate)
		{
			ID:         "pay_ICICI1",
			CreatedAt:  now - 1800,
			PaymentID:  "pay_ICICI1",
			MerchantID: "merchant2",
			Amount:     12000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "icici",
			CapturedAt: now - 1700,
		},
		{
			ID:         "pay_ICICI2",
			CreatedAt:  now - 1200,
			PaymentID:  "pay_ICICI2",
			MerchantID: "merchant2",
			Amount:     18000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "icici",
			CapturedAt: 0,
		},
		{
			ID:         "pay_ICICI3",
			CreatedAt:  now - 600,
			PaymentID:  "pay_ICICI3",
			MerchantID: "merchant2",
			Amount:     22000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "icici",
			CapturedAt: 0,
		},

		// Previous hour's data (all successful to create a drop)
		{
			ID:         "pay_HDFC_OLD1",
			CreatedAt:  now - 3600, // 1 hour ago
			PaymentID:  "pay_HDFC_OLD1",
			MerchantID: "merchant1",
			Amount:     10000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HDFC_OLD2",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_OLD2",
			MerchantID: "merchant1",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_ICICI_OLD1",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_ICICI_OLD1",
			MerchantID: "merchant2",
			Amount:     12000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "icici",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_ICICI_OLD2",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_ICICI_OLD2",
			MerchantID: "merchant2",
			Amount:     18000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "icici",
			CapturedAt: now - 3500,
		},
	}

	// Insert payments
	for _, payment := range payments {
		result := db.Table("payments").Create(&payment)
		if result.Error != nil {
			fmt.Printf("Error inserting payment %s: %v\n", payment.PaymentID, result.Error)
		} else {
			fmt.Printf("Successfully inserted payment %s\n", payment.PaymentID)
		}
	}

	fmt.Println("All alert-triggering payments inserted successfully!")
} 
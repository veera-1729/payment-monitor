package scripts

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InsertNormalPayments(dsn string) {
	// Connect to the databas
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Clear existing payments
	if err := db.Exec("TRUNCATE TABLE payments").Error; err != nil {
		fmt.Printf("Error clearing existing payments: %v\n", err)
		return
	}
	fmt.Println("Cleared existing payments")

	// Get current timestamp
	now := time.Now().Unix()

	// Create payments for different gateways with good success rates
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
		// HDFC payments (all successful)
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
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1100,
		},
		{
			ID:         "pay_HDFC3",
			CreatedAt:  now - 600, // 10 minutes ago
			PaymentID:  "pay_HDFC3",
			MerchantID: "merchant1",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 500,
		},

		// ICICI payments (all successful)
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
			Status:     "STATUS_CAPTURED",
			Gateway:    "icici",
			CapturedAt: now - 1100,
		},
		{
			ID:         "pay_ICICI3",
			CreatedAt:  now - 600,
			PaymentID:  "pay_ICICI3",
			MerchantID: "merchant2",
			Amount:     22000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "icici",
			CapturedAt: now - 500,
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

	fmt.Println("All normal payments inserted successfully!")
}

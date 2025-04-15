package seeder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type Seeder struct {
	db *gorm.DB
}

func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/seed/normal", s.seedNormalPayments)
	mux.HandleFunc("/api/v1/seed/gateway-alert", s.seedGatewayAlertPayments)
	mux.HandleFunc("/api/v1/seed/gateway-method-alert", s.seedGatewayMethodAlertPayments)
	mux.HandleFunc("/api/v1/seed/gateway-merchant-alert", s.seedGatewayMerchantAlertPayments)
	mux.HandleFunc("/api/v1/seed/delete", s.deletePayments)
}

func (s *Seeder) seedNormalPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now().Unix()
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
			CreatedAt:  now - 1800,
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
			CreatedAt:  now - 1200,
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
			CreatedAt:  now - 600,
			PaymentID:  "pay_HDFC3",
			MerchantID: "merchant1",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 500,
		},
	}

	for _, payment := range payments {
		if err := s.db.Table("payments").Create(&payment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Normal payments seeded successfully"})
}

func (s *Seeder) seedGatewayAlertPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now().Unix()
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
		// Current hour data (33% success rate)
		{
			ID:         "pay_HDFC1",
			CreatedAt:  now - 1800,
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
			CreatedAt:  now - 1200,
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
			CreatedAt:  now - 600,
			PaymentID:  "pay_HDFC3",
			MerchantID: "merchant1",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		// Previous hour data (100% success rate)
		{
			ID:         "pay_HDFC_OLD1",
			CreatedAt:  now - 3600,
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
			ID:         "pay_HDFC_OLD3",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_OLD3",
			MerchantID: "merchant1",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		// Add more data points to ensure we meet min_transactions threshold
		{
			ID:         "pay_HDFC4",
			CreatedAt:  now - 900,
			PaymentID:  "pay_HDFC4",
			MerchantID: "merchant1",
			Amount:     25000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HDFC5",
			CreatedAt:  now - 300,
			PaymentID:  "pay_HDFC5",
			MerchantID: "merchant1",
			Amount:     30000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HDFC_OLD4",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_OLD4",
			MerchantID: "merchant1",
			Amount:     25000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HDFC_OLD5",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_OLD5",
			MerchantID: "merchant1",
			Amount:     30000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
	}

	for _, payment := range payments {
		if err := s.db.Table("payments").Create(&payment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Gateway alert payments seeded successfully"})
}

func (s *Seeder) seedGatewayMethodAlertPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now().Unix()
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
		// Current hour data for HDFC UPI (20% success rate)
		{
			ID:         "pay_HUPI1",
			CreatedAt:  now - 1800,
			PaymentID:  "pay_HUPI1",
			MerchantID: "merchant1",
			Amount:     5000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         "pay_HUPI2",
			CreatedAt:  now - 1200,
			PaymentID:  "pay_HUPI2",
			MerchantID: "merchant1",
			Amount:     7000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HUPI3",
			CreatedAt:  now - 600,
			PaymentID:  "pay_HUPI3",
			MerchantID: "merchant1",
			Amount:     9000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HUPI4",
			CreatedAt:  now - 900,
			PaymentID:  "pay_HUPI4",
			MerchantID: "merchant1",
			Amount:     6000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HUPI5",
			CreatedAt:  now - 300,
			PaymentID:  "pay_HUPI5",
			MerchantID: "merchant1",
			Amount:     8000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		// Previous hour data for HDFC UPI (100% success rate)
		{
			ID:         "pay_HUPI_1",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HUPI_1",
			MerchantID: "merchant1",
			Amount:     5000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HUPI_2",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HUPI_2",
			MerchantID: "merchant1",
			Amount:     7000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HUPI_3",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HUPI_3",
			MerchantID: "merchant1",
			Amount:     9000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HUPI_4",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HUPI_4",
			MerchantID: "merchant1",
			Amount:     6000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HUPI_5",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HUPI_5",
			MerchantID: "merchant1",
			Amount:     8000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
	}

	for _, payment := range payments {
		if err := s.db.Table("payments").Create(&payment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Gateway-Method alert payments seeded successfully"})
}

func (s *Seeder) seedGatewayMerchantAlertPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	now := time.Now().Unix()
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
		// Current hour data for HDFC merchant2 (25% success rate)
		{
			ID:         "pay_HDFC_M2_1",
			CreatedAt:  now - 1800,
			PaymentID:  "pay_HDFC_M2_1",
			MerchantID: "merchant2",
			Amount:     12000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         "pay_HDFC_M2_2",
			CreatedAt:  now - 1200,
			PaymentID:  "pay_HDFC_M2_2",
			MerchantID: "merchant2",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HDFC_M2_3",
			CreatedAt:  now - 600,
			PaymentID:  "pay_HDFC_M2_3",
			MerchantID: "merchant2",
			Amount:     18000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HDFC_M2_4",
			CreatedAt:  now - 900,
			PaymentID:  "pay_HDFC_M2_4",
			MerchantID: "merchant2",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         "pay_HDFC_M2_5",
			CreatedAt:  now - 300,
			PaymentID:  "pay_HDFC_M2_5",
			MerchantID: "merchant2",
			Amount:     22000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		// Previous hour data for HDFC merchant2 (100% success rate)
		{
			ID:         "pay_HDFC_M2_OLD1",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_M2_OLD1",
			MerchantID: "merchant2",
			Amount:     12000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HDFC_M2_OLD2",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_M2_OLD2",
			MerchantID: "merchant2",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HDFC_M2_OLD3",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_M2_OLD3",
			MerchantID: "merchant2",
			Amount:     18000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HDFC_M2_OLD4",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_M2_OLD4",
			MerchantID: "merchant2",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         "pay_HDFC_M2_OLD5",
			CreatedAt:  now - 3600,
			PaymentID:  "pay_HDFC_M2_OLD5",
			MerchantID: "merchant2",
			Amount:     22000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
	}

	for _, payment := range payments {
		if err := s.db.Table("payments").Create(&payment).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Gateway-Merchant alert payments seeded successfully"})
}

func (s *Seeder) deletePayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// First, get the count of records to be deleted
	var count int64
	if err := s.db.Raw("SELECT COUNT(*) FROM payments").Scan(&count).Error; err != nil {
		http.Error(w, fmt.Sprintf("Error counting records: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Found %d records to delete\n", count)

	// Delete all records from the payments table using raw SQL
	result := s.db.Exec("TRUNCATE TABLE payments CASCADE")
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("Error deleting records: %v", result.Error), http.StatusInternalServerError)
		return
	}

	// Verify deletion
	var newCount int64
	if err := s.db.Raw("SELECT COUNT(*) FROM payments").Scan(&newCount).Error; err != nil {
		http.Error(w, fmt.Sprintf("Error verifying deletion: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Printf("Records after deletion: %d\n", newCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Payments deleted successfully",
		"count":     count,
		"remaining": newCount,
	})
} 
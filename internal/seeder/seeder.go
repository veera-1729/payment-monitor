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

	// Create a short random suffix for IDs to avoid conflicts
	suffix := fmt.Sprintf("%04d", time.Now().Unix()%10000)

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
			ID:         fmt.Sprintf("pay_HD1_%s", suffix),
			CreatedAt:  now - 1800, // 30 minutes ago
			PaymentID:  fmt.Sprintf("pay_HD1_%s", suffix),
			MerchantID: "merchant1",
			Amount:     10000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         fmt.Sprintf("pay_HD2_%s", suffix),
			CreatedAt:  now - 1200, // 20 minutes ago
			PaymentID:  fmt.Sprintf("pay_HD2_%s", suffix),
			MerchantID: "merchant1",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1100,
		},
		{
			ID:         fmt.Sprintf("pay_HD3_%s", suffix),
			CreatedAt:  now - 600,
			PaymentID:  fmt.Sprintf("pay_HD3_%s", suffix),
			MerchantID: "merchant1",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
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

	// Create a short random suffix for IDs to avoid conflicts
	suffix := fmt.Sprintf("%04d", time.Now().Unix()%10000)

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
			ID:         fmt.Sprintf("pay_HD1_%s", suffix),
			CreatedAt:  now - 1800,
			PaymentID:  fmt.Sprintf("pay_HD1_%s", suffix),
			MerchantID: "merchant1",
			Amount:     10000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         fmt.Sprintf("pay_HD2_%s", suffix),
			CreatedAt:  now - 1200,
			PaymentID:  fmt.Sprintf("pay_HD2_%s", suffix),
			MerchantID: "merchant1",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HD3_%s", suffix),
			CreatedAt:  now - 600,
			PaymentID:  fmt.Sprintf("pay_HD3_%s", suffix),
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
			ID:         fmt.Sprintf("pay_HDO1_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDO1_%s", suffix),
			MerchantID: "merchant1",
			Amount:     10000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDO2_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDO2_%s", suffix),
			MerchantID: "merchant1",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDO3_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDO3_%s", suffix),
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
			ID:         fmt.Sprintf("pay_HD4_%s", suffix),
			CreatedAt:  now - 900,
			PaymentID:  fmt.Sprintf("pay_HD4_%s", suffix),
			MerchantID: "merchant1",
			Amount:     25000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HD5_%s", suffix),
			CreatedAt:  now - 300,
			PaymentID:  fmt.Sprintf("pay_HD5_%s", suffix),
			MerchantID: "merchant1",
			Amount:     30000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HDO4_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDO4_%s", suffix),
			MerchantID: "merchant1",
			Amount:     25000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDO5_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDO5_%s", suffix),
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

	// Create a short random suffix for IDs to avoid conflicts
	suffix := fmt.Sprintf("%04d", time.Now().Unix()%10000)

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
			ID:         fmt.Sprintf("pay_HUPI1_%s", suffix),
			CreatedAt:  now - 1800,
			PaymentID:  fmt.Sprintf("pay_HUPI1_%s", suffix),
			MerchantID: "merchant1",
			Amount:     5000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI2_%s", suffix),
			CreatedAt:  now - 1200,
			PaymentID:  fmt.Sprintf("pay_HUPI2_%s", suffix),
			MerchantID: "merchant1",
			Amount:     7000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI3_%s", suffix),
			CreatedAt:  now - 600,
			PaymentID:  fmt.Sprintf("pay_HUPI3_%s", suffix),
			MerchantID: "merchant1",
			Amount:     9000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI4_%s", suffix),
			CreatedAt:  now - 900,
			PaymentID:  fmt.Sprintf("pay_HUPI4_%s", suffix),
			MerchantID: "merchant1",
			Amount:     6000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI5_%s", suffix),
			CreatedAt:  now - 300,
			PaymentID:  fmt.Sprintf("pay_HUPI5_%s", suffix),
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
			ID:         fmt.Sprintf("pay_HUPI_1_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HUPI_1_%s", suffix),
			MerchantID: "merchant1",
			Amount:     5000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI_2_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HUPI_2_%s", suffix),
			MerchantID: "merchant1",
			Amount:     7000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI_3_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HUPI_3_%s", suffix),
			MerchantID: "merchant1",
			Amount:     9000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI_4_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HUPI_4_%s", suffix),
			MerchantID: "merchant1",
			Amount:     6000,
			Currency:   "INR",
			Method:     "upi",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HUPI_5_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HUPI_5_%s", suffix),
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

	// Create a short random suffix for IDs to avoid conflicts
	suffix := fmt.Sprintf("%04d", time.Now().Unix()%10000)

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
			ID:         fmt.Sprintf("pay_HDFC_M2_1_%s", suffix),
			CreatedAt:  now - 1800,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_1_%s", suffix),
			MerchantID: "merchant2",
			Amount:     12000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 1700,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_2_%s", suffix),
			CreatedAt:  now - 1200,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_2_%s", suffix),
			MerchantID: "merchant2",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_3_%s", suffix),
			CreatedAt:  now - 600,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_3_%s", suffix),
			MerchantID: "merchant2",
			Amount:     18000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_4_%s", suffix),
			CreatedAt:  now - 900,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_4_%s", suffix),
			MerchantID: "merchant2",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_FAILED",
			Gateway:    "hdfc",
			CapturedAt: 0,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_5_%s", suffix),
			CreatedAt:  now - 300,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_5_%s", suffix),
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
			ID:         fmt.Sprintf("pay_HDFC_M2_OLD1_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_OLD1_%s", suffix),
			MerchantID: "merchant2",
			Amount:     12000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_OLD2_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_OLD2_%s", suffix),
			MerchantID: "merchant2",
			Amount:     15000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_OLD3_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_OLD3_%s", suffix),
			MerchantID: "merchant2",
			Amount:     18000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_OLD4_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_OLD4_%s", suffix),
			MerchantID: "merchant2",
			Amount:     20000,
			Currency:   "INR",
			Method:     "card",
			Status:     "STATUS_CAPTURED",
			Gateway:    "hdfc",
			CapturedAt: now - 3500,
		},
		{
			ID:         fmt.Sprintf("pay_HDFC_M2_OLD5_%s", suffix),
			CreatedAt:  now - 3600,
			PaymentID:  fmt.Sprintf("pay_HDFC_M2_OLD5_%s", suffix),
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

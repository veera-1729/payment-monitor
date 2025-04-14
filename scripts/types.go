package scripts

type Payment struct {
	CreatedAt      int64  `gorm:"column:created_at"`
	PaymentID      string `gorm:"column:payment_id"`
	MerchantID     string `gorm:"column:merchant_id"`
	Amount         int64  `gorm:"column:amount"`
	Currency       string `gorm:"column:currency"`
	Method         string `gorm:"column:method"`
	Status         string `gorm:"column:status"`
	Gateway        string `gorm:"column:gateway"`
	CapturedAt     int64  `gorm:"column:captured_at"`
} 
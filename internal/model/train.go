package model

import "time"

// Train 对应 MySQL 的 trains 表
type Train struct {
	ID string `gorm:"column:id;primaryKey;size:32" json:"id"`

	FromStation string `gorm:"column:from_station;size:64;not null" json:"from_station"`
	ToStation   string `gorm:"column:to_station;size:64;not null" json:"to_station"`

	// 单位为分，不使用 float64 保存票价
	PriceCents int64 `gorm:"column:price_cents;not null" json:"price_cents"`

	TotalSeats     int `gorm:"column:total_seats;not null" json:"total_seats"`
	AvailableSeats int `gorm:"column:available_seats;not null" json:"available_seats"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Train) TableName() string {
	return "trains"
}

// PurchaseLog 是本节的扣票流水，不是完整订单
type PurchaseLog struct {
	ID uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	TrainID  string `gorm:"column:train_id;size:32;not null" json:"train_id"`
	Quantity int    `gorm:"column:quantity;not null" json:"quantity"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PurchaseLog) TableName() string {
	return "purchase_logs"
}

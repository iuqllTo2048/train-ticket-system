package model

import "time"

// 订单状态常量
const (
	OrderStatusSuccess   = "SUCCESS"
	OrderStatusCancelled = "CANCELLED"
)

// Order 对应 MySQL 中的 orders 表。
type Order struct {
	ID uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`

	OrderNo string `gorm:"column:order_no;size:64;not null;uniqueIndex" json:"order_no"`

	TrainID string `gorm:"column:train_id;size:32;not null" json:"train_id"`

	UserID string `gorm:"column:user_id;size:64;not null" json:"user_id"`

	Quantity int `gorm:"column:quantity;not null" json:"quantity"`

	UnitPriceCents int64 `gorm:"column:unit_price_cents;not null" json:"unit_price_cents"`

	TotalAmountCents int64 `gorm:"column:total_amount_cents;not null" json:"total_amount_cents"`

	Status string `gorm:"column:status;size:16;not null" json:"status"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`

	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Order) TableName() string {
	return "orders"
}

package domain

import "time"

var orderTransitions = map[OrderStatus][]OrderStatus{
	OrderNew:       {OrderContacted, OrderCancelled},
	OrderContacted: {OrderConfirmed, OrderCancelled},
	OrderConfirmed: {OrderCompleted, OrderCancelled},
	OrderCancelled: {},
	OrderCompleted: {},
}

type Order struct {
	ID               int64       `gorm:"primaryKey;autoIncrement"`
	VariantID        int64       `gorm:"not null"`
	CustomerName     string      `gorm:"not null"`
	TelegramUsername string      `gorm:"not null"`
	Phone            *string
	Comment          *string
	Status           OrderStatus `gorm:"not null;default:'NEW'"`
	TgNotifiedAt     *time.Time
	CreatedAt        time.Time   `gorm:"not null;default:now()"`
	UpdatedAt        time.Time   `gorm:"not null;default:now()"`

	Variant ProductVariant `gorm:"foreignKey:VariantID"`
}

func (Order) TableName() string { return "orders" }

func (o *Order) CanTransitionTo(next OrderStatus) bool {
	for _, s := range orderTransitions[o.Status] {
		if s == next { return true }
	}
	return false
}

func (o *Order) AllowedTransitions() []OrderStatus {
	return orderTransitions[o.Status]
}

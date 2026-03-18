package domain

import "time"

// orderTransitions defines the allowed FSM transitions.
var orderTransitions = map[OrderStatus][]OrderStatus{
	OrderNew:       {OrderContacted, OrderCancelled},
	OrderContacted: {OrderConfirmed, OrderCancelled},
	OrderConfirmed: {OrderCompleted, OrderCancelled},
	OrderCancelled: {},
	OrderCompleted: {},
}

type Order struct {
	ID               string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	VariantID        string      `gorm:"not null"`
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

// CanTransitionTo checks if the FSM allows moving to the next status.
func (o *Order) CanTransitionTo(next OrderStatus) bool {
	allowed := orderTransitions[o.Status]
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// AllowedTransitions returns the valid next statuses from current state.
func (o *Order) AllowedTransitions() []OrderStatus {
	return orderTransitions[o.Status]
}

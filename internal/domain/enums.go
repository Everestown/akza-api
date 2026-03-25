package domain

type CollectionStatus string

const (
	CollectionDraft     CollectionStatus = "DRAFT"
	CollectionScheduled CollectionStatus = "SCHEDULED"
	CollectionPublished CollectionStatus = "PUBLISHED"
	CollectionArchived  CollectionStatus = "ARCHIVED"
)

func (s CollectionStatus) IsValid() bool {
	switch s {
	case CollectionDraft, CollectionScheduled, CollectionPublished, CollectionArchived:
		return true
	}
	return false
}

type OrderStatus string

const (
	OrderNew       OrderStatus = "NEW"
	OrderContacted OrderStatus = "CONTACTED"
	OrderConfirmed OrderStatus = "CONFIRMED"
	OrderCancelled OrderStatus = "CANCELLED"
	OrderCompleted OrderStatus = "COMPLETED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderNew, OrderContacted, OrderConfirmed, OrderCancelled, OrderCompleted:
		return true
	}
	return false
}

type MediaType string

const (
	MediaImage MediaType = "IMAGE"
	MediaVideo MediaType = "VIDEO"
)

func (t MediaType) IsValid() bool { return t == MediaImage || t == MediaVideo }

type PageSection string

const (
	SectionHero     PageSection = "HERO"
	SectionAbout    PageSection = "ABOUT"
	SectionContacts PageSection = "CONTACTS"
	SectionFooter   PageSection = "FOOTER"
	SectionHeader     PageSection = "HEADER"
	SectionDictionary PageSection = "DICTIONARY"
)

func (s PageSection) IsValid() bool {
	switch s {
	case SectionHero, SectionAbout, SectionContacts, SectionFooter, SectionHeader, SectionDictionary:
		return true
	}
	return false
}

package entity

// CategoryConcept — unifies same categories in different languages
type CategoryConcept struct {
	ID          int
	Key         string // key in database
	Description string // readable
	I18ns       []*CategoryI18n
}

// CategoryI18n — represents category in specified language
type CategoryI18n struct {
	ConceptID int
	Lang      string // ru / en / de ...
	Title     string // official name in wiki without 'Category:' prefix
	Name      string // readable name
}

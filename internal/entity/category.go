package entity

// CategoryConcept — концепт категории
type CategoryConcept struct {
	ID          int
	Key         string
	Description string
	I18ns       []*CategoryI18n
}

// CategoryI18n — локализация концепта
type CategoryI18n struct {
	ConceptID int
	Lang      string
	Title     string
	Name      string
}

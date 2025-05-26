package entity

import "errors"

var ErrCategoryNotFound = errors.New("category not found")

type Category string

// todo shitty id, i think we dont have to take care or let it for translations api
//// CategoryConcept — unifies same categories in different languages
//type CategoryConcept struct {
//	Lang          string   // ru / en / de ...
//	Category      Category // official name in wiki without 'Category:' prefix
//	Subcategories []Category
//}
//
//// CategoryI18n — represents category in specified language
//type CategoryI18n struct {
//	ConceptID int
//	Lang      string // ru / en / de ...
//	Category  string // official name in wiki without 'Category:' prefix
//	Name      string // readable name
//}

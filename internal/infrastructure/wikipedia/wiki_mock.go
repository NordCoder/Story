package wikipedia

import (
	"context"
	"fmt"
)

type wikiMock struct{}

// NewWikiMock создаёт новый экземпляр заглушки
func NewWikiMock() WikiClient {
	return &wikiMock{}
}

// GetCategorySummaries отдаёт до limit статей для заданной категории.
// Поддерживаем три популярных категории: Вторая мировая война, Go (язык), Машинное обучение.
func (w *wikiMock) GetCategorySummaries(ctx context.Context, category string, limit int) ([]*ArticleSummary, error) {
	var items []*ArticleSummary

	switch category {
	case "World_War_II", "Вторая_мировая_война":
		items = []*ArticleSummary{
			{
				Title:    "Battle of Stalingrad",
				Extract:  "The Battle of Stalingrad was a major battle on the Eastern Front of World War II in which Nazi Germany and its allies fought the Soviet Union for control of the city of Stalingrad.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/a/a3/Stamp_of_Russia-2001-2004-198.jpg",
				PageURL:  "https://en.wikipedia.org/wiki/Battle_of_Stalingrad",
			},
			{
				Title:    "D-Day",
				Extract:  "D-Day was the Allied invasion of Normandy on 6 June 1944. It was one of the largest amphibious military assaults in history and began the liberation of German-occupied Western Europe.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/a/a2/Omaha_beach_casualties.jpg",
				PageURL:  "https://en.wikipedia.org/wiki/Normandy_landings",
			},
			{
				Title:    "Operation Barbarossa",
				Extract:  "Operation Barbarossa was the code name for the Axis invasion of the Soviet Union, which started on 22 June 1941 and marked the beginning of the largest theatre of war in history.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/8/8e/Operation_Barbarossa_%28map%29.svg",
				PageURL:  "https://en.wikipedia.org/wiki/Operation_Barbarossa",
			},
			{
				Title:    "Holocaust",
				Extract:  "The Holocaust was the genocide of European Jews during World War II, in which Nazi Germany and its collaborators systematically murdered six million Jews.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/5/5c/Library_of_Congress_91115u.jpg",
				PageURL:  "https://en.wikipedia.org/wiki/The_Holocaust",
			},
			{
				Title:    "Battle of Kursk",
				Extract:  "The Battle of Kursk was a Second World War engagement between German and Soviet forces on the Eastern Front near Kursk in the Soviet Union.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/9/95/Kursk_tank_battle_map.png",
				PageURL:  "https://en.wikipedia.org/wiki/Battle_of_Kursk",
			},
		}

	case "Go_(programming_language)":
		items = []*ArticleSummary{
			{
				Title:    "Go (programming language)",
				Extract:  "Go is a statically typed, compiled programming language designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/0/05/Go_Logo_Blue.svg",
				PageURL:  "https://en.wikipedia.org/wiki/Go_(programming_language)",
			},
			{
				Title:    "Goroutine",
				Extract:  "A goroutine is a lightweight thread managed by the Go runtime.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/f/f2/Gopher_pre_winter.jpg",
				PageURL:  "https://en.wikipedia.org/wiki/Goroutine",
			},
			{
				Title:    "Channels (Go)",
				Extract:  "Channels are a typed conduit through which you can send and receive values with the channel operator, <-.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/1/19/Channels.svg",
				PageURL:  "https://en.wikipedia.org/wiki/Channel_(programming)",
			},
			{
				Title:    "Interfaces in Go",
				Extract:  "An interface type is defined by a set of method signatures and describes a set of methods that a type must have to implement the interface.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/4/4f/Go_interfaces.svg",
				PageURL:  "https://en.wikipedia.org/wiki/Interface_(Go)",
			},
		}

	case "Machine_learning":
		items = []*ArticleSummary{
			{
				Title:    "Machine learning",
				Extract:  "Machine learning is the study of computer algorithms that improve automatically through experience.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/4/44/Machine_learning.svg",
				PageURL:  "https://en.wikipedia.org/wiki/Machine_learning",
			},
			{
				Title:    "Supervised learning",
				Extract:  "Supervised learning is the machine learning task of learning a function that maps an input to an output based on example input–output pairs.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/7/77/Supervised_learning_overview.png",
				PageURL:  "https://en.wikipedia.org/wiki/Supervised_learning",
			},
			{
				Title:    "Neural network",
				Extract:  "In machine learning, a neural network or artificial neural network is a network of artificial neurons modeled after biological neural networks.",
				ImageURL: "https://upload.wikimedia.org/wikipedia/commons/6/60/Artificial_neural_network.svg",
				PageURL:  "https://en.wikipedia.org/wiki/Artificial_neural_network",
			},
		}

	default:
		return nil, fmt.Errorf("unknown category %q", category)
	}

	// Применяем limit
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}
	return items, nil
}

func (w *wikiMock) Ping(ctx context.Context) error {
	return nil
}

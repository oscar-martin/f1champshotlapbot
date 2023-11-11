package tracks

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

type Category struct {
	ID       string
	Name     string
	Sessions []Session
}

func (c Category) CommandString(trackId string) string {
	return " ▸ " + c.Name + fmt.Sprintf(" ➡ /%s_%s", trackId, c.ID)
}

type Track struct {
	Command    string
	ID         string
	Name       string
	Categories []Category
	mu         sync.Mutex
}

func (t *Track) GetCategories(ctx context.Context, domain string) ([]Category, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.Categories) == 0 {
		// if there is no categories, fetch them
		ss, err := getSessions(ctx, t.Name, domain)
		if err != nil {
			return nil, err
		}
		t.Categories = getCategories(ss)
	}

	return t.Categories, nil
}

func (t *Track) GetCategoryById(cId string) (Category, bool) {
	for _, c := range t.Categories {
		if c.ID == cId {
			return c, true
		}
	}
	return Category{}, false
}

func (t *Track) CommandString() string {
	return " ▸ " + t.Name + " ➡ " + t.Command
}

func getCategories(ss []Session) []Category {
	cats := map[string]Category{}
	for _, session := range ss {
		id, name := ExtractCategory(session.Category)
		if c, exits := cats[id]; !exits {
			cats[id] = Category{
				ID:       id,
				Name:     name,
				Sessions: []Session{session},
			}
		} else {
			c.Sessions = append(c.Sessions, session)
			cats[id] = c
		}
	}

	categories := make([]Category, 0, len(cats))

	for _, c := range cats {
		categories = append(categories, c)
	}

	// sort categories by name
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories
}

func ExtractCategory(category string) (id string, name string) {
	id = category
	name = category
	if len(category) > 0 {
		name = strings.Split(category, ",")[0]
		id = strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	}
	return
}

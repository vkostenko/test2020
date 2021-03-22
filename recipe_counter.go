package main

import (
	"sort"
	"strings"
	"sync"
)

type recipeCounter struct {
	mx             *sync.RWMutex
	data           map[string]uint32
	searchRecipes  []string
	matchedRecipes []string
}

func newRecipeCounter(searchRecipes []string) *recipeCounter {
	return &recipeCounter{
		mx:            &sync.RWMutex{},
		data:          map[string]uint32{},
		searchRecipes: searchRecipes,
	}
}

func (r *recipeCounter) add(item inputItem) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.data[item.Recipe]; !ok {
		r.data[item.Recipe] = 1
		if containsOneOf(item.Recipe, r.searchRecipes) {
			r.matchedRecipes = append(r.matchedRecipes, item.Recipe)
		}

		return
	}

	r.data[item.Recipe]++
}

func containsOneOf(name string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(name, substr) {
			return true
		}
	}

	return false
}

func (r *recipeCounter) getUniqueRecipesCount() uint32 {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return uint32(len(r.data))
}

func (r *recipeCounter) getCountByRecipes() []RecipeCount {
	r.mx.RLock()
	defer r.mx.RUnlock()

	keys := mapKeysToSlice(r.data)
	sortStringSlice(keys)

	result := make([]RecipeCount, len(r.data))
	for i, key := range keys {
		result[i] = RecipeCount{
			Recipe: key,
			Count:  r.data[key],
		}
	}

	return result
}

func mapKeysToSlice(input map[string]uint32) []string {
	result := make([]string, 0, len(input))
	for key := range input {
		result = append(result, key)
	}

	return result
}

func (r *recipeCounter) getMatchedRecipes() []string {
	r.mx.Lock()
	defer r.mx.Unlock()

	sortStringSlice(r.matchedRecipes)

	return r.matchedRecipes
}

func sortStringSlice(input []string) {
	sort.Slice(input, func(i, j int) bool {
		return input[i] < input[j]
	})
}

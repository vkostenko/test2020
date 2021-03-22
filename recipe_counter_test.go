package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRecipeCounter(t *testing.T) {
	data := []string{
		"a", "a", "b", "c", "a", "ad", "b", "aa",
	}
	expectedCount := []RecipeCount{
		{
			Recipe: "a",
			Count:  3,
		},
		{
			Recipe: "aa",
			Count:  1,
		},
		{
			Recipe: "ad",
			Count:  1,
		},
		{
			Recipe: "b",
			Count:  2,
		},
		{
			Recipe: "c",
			Count:  1,
		},
	}

	searchedRecipes := []string{"a", "d"}
	expectedMatches := []string{"a", "aa", "ad"}

	counter := newRecipeCounter(searchedRecipes)
	for _, name := range data {
		inputItem := inputItem{
			Recipe: name,
		}
		counter.add(inputItem)
	}

	require.Equal(t, expectedCount, counter.getCountByRecipes())

	require.Equal(t, expectedMatches, counter.getMatchedRecipes())
}

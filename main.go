package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

type inputItem struct {
	Postcode string `json:"postcode"`
	Recipe   string `json:"recipe"`
	Delivery string `json:"delivery"`
}

type result struct {
	UniqueRecipeCount       uint32                  `json:"unique_recipe_count"`
	CountPerRecipe          []RecipeCount           `json:"count_per_recipe"`
	BusiestPostcode         BusiestPostcode         `json:"busiest_postcode"`
	CountPerPostcodeAndTime CountPerPostcodeAndTime `json:"count_per_postcode_and_time"`
	MatchByName             []string                `json:"match_by_name"`
}

type RecipeCount struct {
	Recipe string `json:"recipe"`
	Count  uint32 `json:"count"`
}

type BusiestPostcode struct {
	Postcode      string `json:"postcode"`
	DeliveryCount uint32 `json:"delivery_count"`
}

type CountPerPostcodeAndTime struct {
	Postcode      string `json:"postcode"`
	From          string `json:"from"`
	To            string `json:"to"`
	DeliveryCount uint32 `json:"delivery_count"`
}

func main() {
	inputFilePath := flag.String("input_file", "", "Path to input file. Should be in JSON format")
	searchRecipeNames := flag.String("recipe_names", "", "Recipe names to search joined by comma or another delimiter if provided")
	searchRecipeNamesDelimiter := flag.String("recipe_names_delimiter", ",", "Delimiter for recipe names to search")
	countDeliveriesByPostcode := flag.String("deliveries_by_postcode", "10120", "count deliveries in JSON file with postcode")
	countDeliveriesByPostcodeFromTime := flag.String("deliveries_by_postcode_from_time", "11AM", "count deliveries in JSON file with postcode after provided time")
	countDeliveriesByPostcodeToTime := flag.String("deliveries_by_postcode_to_time", "3PM", "count deliveries in JSON file with postcode until provided time")
	flag.Parse()

	if inputFilePath == nil || *inputFilePath == "" {
		printErr("Mandatory flag not provided: input_file")
		return
	}

	inputFile, err := os.Open(*inputFilePath)
	if err != nil {
		printErr(fmt.Sprintf("error opening file %s: %v", inputFilePath, err))
		return
	}

	recipeCounter := newRecipeCounter(strings.Split(*searchRecipeNames, *searchRecipeNamesDelimiter))
	postcodeCounter := newPostcodeCounter()
	postcodeTimeMatcher, err := newPostcodeTimeMatcher(*countDeliveriesByPostcode, *countDeliveriesByPostcodeFromTime, *countDeliveriesByPostcodeToTime)
	if err != nil {
		printErr(fmt.Sprintf("Failed to init postcodeTimeMatcher %v", err))
		return
	}

	deliveryItems := make(chan inputItem, 1)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	recoveredGo(func() {
		defer wg.Done()
		readJsonStream(inputFile, deliveryItems)
	})

	recoveredGo(func() {
		defer wg.Done()
		applyDeliveryItems(recipeCounter, postcodeCounter, postcodeTimeMatcher, deliveryItems)
	})

	wg.Wait()

	result := result{
		UniqueRecipeCount: recipeCounter.getUniqueRecipesCount(),
		CountPerRecipe:    recipeCounter.getCountByRecipes(),
		BusiestPostcode:   postcodeCounter.getBusiestPostcode(),
		CountPerPostcodeAndTime: CountPerPostcodeAndTime{
			Postcode:      *countDeliveriesByPostcode,
			From:          *countDeliveriesByPostcodeFromTime,
			To:            *countDeliveriesByPostcodeToTime,
			DeliveryCount: postcodeTimeMatcher.getMatchedItemsCount(),
		},
		MatchByName: recipeCounter.getMatchedRecipes(),
	}

	msg, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		printErr(fmt.Sprintf("error marshalling: %v", err))
		return
	}

	fmt.Fprint(os.Stdout, string(msg))
}

func applyDeliveryItems(recipeCounter *recipeCounter, postcodeCounter *postcodeCounter, matcher *postcodeTimeMatcher, deliveryItems <-chan inputItem) {
	wg := &sync.WaitGroup{}
	for deliveryItem := range deliveryItems {
		wg.Add(3)

		deliveryItem := deliveryItem
		recoveredGo(func() {
			defer wg.Done()
			recipeCounter.add(deliveryItem)
		})

		recoveredGo(func() {
			defer wg.Done()
			postcodeCounter.add(deliveryItem)
		})

		recoveredGo(func() {
			defer wg.Done()
			matcher.tryAdd(deliveryItem)
		})
	}

	wg.Wait()
}

func readJsonStream(reader io.Reader, deliveryItems chan<- inputItem) {
	defer close(deliveryItems)
	decoder := json.NewDecoder(reader)

	token, err := decoder.Token()
	if err != nil {
		printErr(fmt.Sprintf("token %v start call error: %v", token, err))
		return
	}

	for decoder.More() {
		var item inputItem

		err := decoder.Decode(&item)
		if err != nil {
			printErr(fmt.Sprintf("error decoding item: %v", err))
			return
		}

		deliveryItems <- item
	}

	token, err = decoder.Token()
	if err != nil {
		printErr(fmt.Sprintf("token %v end call error: %v", token, err))
	}
}

func recoveredGo(method func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				printErr(fmt.Sprintf("goroutine panic recovered, error: %+v, %s", r, debug.Stack()))
			}
		}()

		method()
	}()
}

func printErr(msg string) {
	fmt.Fprintf(os.Stderr, msg)
}

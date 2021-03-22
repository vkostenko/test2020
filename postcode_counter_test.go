package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostCodeCounter(t *testing.T) {
	data := []string{
		"a", "a", "b", "c", "a", "ad", "b", "aa",
	}
	expectedMaxProduct := "a"
	expectedMaxCount := uint32(3)

	counter := newPostcodeCounter()
	for _, name := range data {
		inputItem := inputItem{
			Postcode: name,
		}
		counter.add(inputItem)
	}

	busiest := counter.getBusiestPostcode()

	require.Equal(t, expectedMaxProduct, busiest.Postcode)
	require.Equal(t, expectedMaxCount, busiest.DeliveryCount)
}

package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestTo24Hour(t *testing.T) {
	testCases := []struct {
		input      string
		output     int8
		errMessage string
	}{
		{"1PM", 13, ""},
		{"1AM", 1, ""},
		{"12PM", 12, ""},
		{"12AM", 0, ""},
		{"0AM", 0, "wrong format 0AM"},
		{"-1AM", 0, "wrong format -1AM"},
		{"13AM", 0, "wrong format 13AM"},
		{"10DM", 0, "wrong format 10DM"},
		{"", 0, "wrong format "},
		{"ABC", 0, "wrong format ABC"},
	}

	as := assert.New(t)
	for i, testCase := range testCases {
		out, err := parseHour12to24(testCase.input)
		as.Equal(testCase.output, out, "element %d", i)
		if testCase.errMessage != "" {
			as.Error(err, "element %d", i)
			as.Equal(testCase.errMessage, err.Error(), "element %d", i)
		}
	}
}

func TestTryAdd(t *testing.T) {
	items := []inputItem{
		{Postcode: "A", Delivery: "Monday 10AM - 3PM"}, // true
		{Postcode: "A", Delivery: "Monday 1PM - 2PM"},  // true
		{Postcode: "A", Delivery: "Monday 10PM - 12AM"},
		{Postcode: "A", Delivery: "Tuesday 9AM - 11AM"}, // true
		{Postcode: "A", Delivery: "Tuesday 8AM - 11AM"},
		{Postcode: "A", Delivery: "Tuesday 3PM - 5PM"},
		{Postcode: "B", Delivery: "Monday 10AM - 11AM"},
	}

	matcher, err := newPostcodeTimeMatcher("A", "9AM", "4PM")
	require.NoError(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(len(items))
	for _, item := range items {
		go func(item inputItem) {
			defer wg.Done()
			matcher.tryAdd(item)
		}(item)
	}
	wg.Wait()

	require.Equal(t, uint32(3), matcher.getMatchedItemsCount())
}

func TestTimeMatched(t *testing.T) {
	t.Run("simple_case", func(t *testing.T) {
		matcher, err := newPostcodeTimeMatcher("A", "2AM", "5AM")
		require.NoError(t, err)

		as := assert.New(t)
		as.True(matcher.timeMatched(3, 4))
		as.True(matcher.timeMatched(2, 5))
		as.False(matcher.timeMatched(2, 6))
		as.False(matcher.timeMatched(1, 5))
		as.False(matcher.timeMatched(1, 7))
		as.False(matcher.timeMatched(23, 1))
		as.False(matcher.timeMatched(23, 4))
	})

	t.Run("case_with_midnight", func(t *testing.T) {
		matcher, err := newPostcodeTimeMatcher("A", "10PM", "3AM")
		require.NoError(t, err)

		as := assert.New(t)
		as.True(matcher.timeMatched(22, 23))
		as.True(matcher.timeMatched(23, 1))
		as.True(matcher.timeMatched(1, 2))
		as.False(matcher.timeMatched(1, 23))
		as.False(matcher.timeMatched(1, 4))
	})
}

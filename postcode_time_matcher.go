package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
)

const (
	hour12Layout       = `^(1[0-2]|0?[1-9])(AM|PM)$`
	deliveryTimeLayout = `^\w+ (1[0-2]|0?[1-9])(AM|PM) - (1[0-2]|0?[1-9])(AM|PM)$`
)

type postcodeTimeMatcher struct {
	counter            uint32
	criteria           searchCriteria
	deliveryTimeRegexp *regexp.Regexp
}

type searchCriteria struct {
	name          string
	startHour     int8
	endHour       int8
	coverMidnight bool
}

func newPostcodeTimeMatcher(name string, inputStartHour string, inputEndHour string) (*postcodeTimeMatcher, error) {
	startHour, err := parseHour12to24(inputStartHour)
	if err != nil {
		return nil, err
	}

	endHour, err := parseHour12to24(inputEndHour)
	if err != nil {
		return nil, err
	}

	return &postcodeTimeMatcher{
		counter:            0,
		deliveryTimeRegexp: regexp.MustCompile(deliveryTimeLayout),
		criteria: searchCriteria{
			name:          name,
			startHour:     startHour,
			endHour:       endHour,
			coverMidnight: startHour > endHour,
		},
	}, nil
}

func parseHour12to24(inputHour string) (int8, error) {
	hour12, pmPart, err := parseHour12(inputHour)
	if err != nil {
		return 0, err
	}

	return toHour24(hour12, pmPart)
}

func parseHour12(inputHour string) (hour, pmPart string, err error) {
	re := regexp.MustCompile(hour12Layout)
	result := re.FindAllStringSubmatch(inputHour, -1)

	if len(result) != 1 || len(result[0]) != 3 {
		return "", "", fmt.Errorf("wrong format %s", inputHour)
	}

	return result[0][1], result[0][2], nil
}

func toHour24(hour12, pmPart string) (int8, error) {
	hour, err := strconv.ParseInt(hour12, 10, 8)
	if err != nil {
		return 0, err
	}

	if hour == 12 {
		hour = 0
	}

	if pmPart == "PM" {
		hour += 12
	}

	return int8(hour), nil
}

func (p *postcodeTimeMatcher) tryAdd(item inputItem) {
	if item.Postcode != p.criteria.name {
		return
	}

	parsed := p.deliveryTimeRegexp.FindAllStringSubmatch(item.Delivery, -1)
	if len(parsed) != 1 || len(parsed[0]) != 5 {
		printErr(fmt.Sprintf("wrong format %s", item.Delivery))
		return
	}

	startHour, err := toHour24(parsed[0][1], parsed[0][2])
	if err != nil {
		printErr(fmt.Sprintf("error converting startHour to hour24 format %s, %v", item.Delivery, err))
	}

	endHour, err := toHour24(parsed[0][3], parsed[0][4])
	if err != nil {
		printErr(fmt.Sprintf("error converting endHour to hour24 format %s, %v", item.Delivery, err))
	}

	if p.timeMatched(startHour, endHour) {
		atomic.AddUint32(&p.counter, 1)
	}
}

func (p *postcodeTimeMatcher) timeMatched(startHour, endHour int8) bool {
	if !p.criteria.coverMidnight {
		if startHour <= endHour {
			return startHour >= p.criteria.startHour && endHour <= p.criteria.endHour
		}

		return false
	}

	if startHour <= endHour {
		return startHour >= p.criteria.startHour || endHour <= p.criteria.endHour
	}

	return startHour >= p.criteria.startHour && endHour <= p.criteria.endHour
}

func (p *postcodeTimeMatcher) getMatchedItemsCount() uint32 {
	return atomic.LoadUint32(&p.counter)
}

package main

import (
	"sync"
)

type postcodeCounter struct {
	mx   *sync.RWMutex
	data map[string]uint32
}

func newPostcodeCounter() *postcodeCounter {
	return &postcodeCounter{
		mx:   &sync.RWMutex{},
		data: map[string]uint32{},
	}
}

func (r *postcodeCounter) add(item inputItem) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.data[item.Postcode]; !ok {
		r.data[item.Postcode] = 1
		return
	}

	r.data[item.Postcode]++
}

func (r *postcodeCounter) getBusiestPostcode() BusiestPostcode {
	r.mx.RLock()
	defer r.mx.RUnlock()

	maxPostCode := ""
	maxValue := uint32(0)

	for postcode, count := range r.data {
		if count > maxValue {
			maxValue = count
			maxPostCode = postcode
		}
	}

	return BusiestPostcode{
		Postcode:      maxPostCode,
		DeliveryCount: maxValue,
	}
}

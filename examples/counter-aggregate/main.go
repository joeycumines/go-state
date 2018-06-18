/*
   Copyright 2018 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

package main

import (
	"github.com/joeycumines/go-bigbuff"
	"github.com/joeycumines/go-state/examples/counter/model"
	"github.com/joeycumines/go-state"
	"context"
	"sync"
	"encoding/json"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buffer := new(bigbuff.Buffer)
	defer buffer.Close()

	consumer, err := buffer.NewConsumer()

	if err != nil {
		panic(err)
	}

	defer consumer.Close()

	newFetcher := func() func(ctx context.Context) (messages []interface{}, ok bool, err error) {
		var (
			fetcherMessages = make(chan []interface{})
			fetcherDone     = make(chan struct{})
		)

		go func() {
			defer close(fetcherDone)

			fetcherMessages <- []interface{}{
				// does nothing the first time, but the second time it has the previous multis
				model.TickTicker{},
				model.TickTicker{},

				// these always adjust by the same amounts
				model.SetHorizontalMultiTicker(1),
				model.SetVerticalMultiTicker(-1),
				model.TickTicker{},
				model.SetHorizontalMultiTicker(-10),
				model.SetVerticalMultiTicker(10),
				model.TickTicker{},
			}
		}()

		return func(ctx context.Context) (messages []interface{}, ok bool, err error) {
			select {
			case <-fetcherDone:
				ok = false
			case messages = <-fetcherMessages:
				ok = true
			}
			return
		}
	}

	store := new(memStore)

	hydrator := func(ctx context.Context, key string, value []byte) (interface{}, []func() (key string, load func(model interface{})), error) {
		// the hydrator converts the raw data to a usable type
		var ticker model.Ticker
		if err := json.Unmarshal(value, &ticker); err != nil {
			return nil, nil, err
		}
		return ticker, nil, nil
	}

	dehydrator := func(ctx context.Context, key string, model interface{}) (value []byte, err error) {
		// the dehydrator reverses the hydrator
		return json.Marshal(model)
	}

	fmt.Println("\nRunning for the first time for 'model_key'...")

	err = state.Aggregate(
		ctx,
		model.InitTicker,
		model.UpdateTicker,
		model.ViewTicker,
		buffer,
		consumer,
		newFetcher(),
		store,
		hydrator,
		dehydrator,
		"model_key",
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("\nMODEL:" + store.loadString("model_key"))

	fmt.Println("\nRunning again against 'model_key'...")

	err = state.Aggregate(
		ctx,
		model.InitTicker,
		model.UpdateTicker,
		model.ViewTicker,
		buffer,
		consumer,
		newFetcher(),
		store,
		hydrator,
		dehydrator,
		"model_key",
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("\nMODEL:" + store.loadString("model_key"))
}

type memStore sync.Map

func (s *memStore) loadString(key string) string {
	m := (*sync.Map)(s)
	v, _ := m.Load(key)
	b, _ := v.([]byte)
	return string(b)
}

func (s *memStore) Load(ctx context.Context, key string) ([]byte, bool, error) {
	m := (*sync.Map)(s)
	v, ok := m.Load(key)
	b, _ := v.([]byte)
	return b, ok, nil
}

func (s *memStore) Store(ctx context.Context, key string, value []byte) error {
	m := (*sync.Map)(s)
	m.Store(key, value)
	return nil
}
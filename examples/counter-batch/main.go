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
	"context"
	"github.com/joeycumines/go-bigbuff"
	"github.com/joeycumines/go-state"
	"github.com/joeycumines/go-state/examples/counter/model"
	"time"
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

	var (
		fetcherMessages = make(chan []interface{})
		fetcherDone     = make(chan struct{})
	)

	go func() {
		defer close(fetcherDone)

		fetcherMessages <- []interface{}{
			model.TickerHorizontalCounter(model.IncrementCounter{}),
			model.TickerHorizontalCounter(model.IncrementCounter{}),
			model.TickerVerticalCounter(model.DecrementCounter{}),
			model.TickerHorizontalCounter(model.DecrementCounter{}),
			model.SetHorizontalMultiTicker(1),
			model.SetVerticalMultiTicker(-1),
			model.TickTicker{},
			model.TickTicker{},
			model.TickTicker{},
			model.SetHorizontalMultiTicker(-10),
			model.SetVerticalMultiTicker(10),
			model.TickTicker{},
			model.TickTicker{},
		}

		time.Sleep(time.Second)

		fetcherMessages <- []interface{}{
			model.SetHorizontalMultiTicker(1),
			model.SetVerticalMultiTicker(-1),
		}
		fetcherMessages <- []interface{}{
			model.TickTicker{},
			model.TickTicker{},
		}

		time.Sleep(time.Second)

		fetcherMessages <- []interface{}{
			model.SetHorizontalMultiTicker(-1),
			model.SetVerticalMultiTicker(1),
		}
		fetcherMessages <- []interface{}{
			model.TickTicker{},
			model.TickTicker{},
		}
		fetcherMessages <- []interface{}{}

		time.Sleep(time.Second)

		fetcherMessages <- []interface{}{}
		fetcherMessages <- []interface{}{
			model.TickerHorizontalCounter(model.DecrementCounter{}),
			model.TickerHorizontalCounter(model.DecrementCounter{}),
			model.TickerVerticalCounter(model.IncrementCounter{}),
			model.TickerHorizontalCounter(model.IncrementCounter{}),
			model.SetHorizontalMultiTicker(-1),
			model.SetVerticalMultiTicker(1),
			model.TickTicker{},
			model.TickTicker{},
			model.TickTicker{},
			model.SetHorizontalMultiTicker(10),
			model.SetVerticalMultiTicker(-10),
			model.TickTicker{},
			model.TickTicker{},
		}
	}()

	err = state.Batch(
		ctx,
		model.InitTicker,
		model.UpdateTicker,
		model.ViewTicker,
		buffer,
		consumer,
		func(ctx context.Context) (messages []interface{}, ok bool, err error) {
			select {
			case <-fetcherDone:
				ok = false
			case messages = <-fetcherMessages:
				ok = true
			}
			return
		},
	)

	if err != nil {
		panic(err)
	}
}

/*
   Copyright 2019 Joseph Cumines

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
	"fmt"
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

	runErr := make(chan error)

	go func() {
		runErr <- state.Run(
			ctx,
			model.InitTicker,
			model.UpdateTicker,
			model.ViewTicker,
			buffer,
			consumer,
		)
	}()

	if err := buffer.Put(
		ctx,
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
	); err != nil {
		panic(err)
	}

	for buffer.Size() != 0 {
		select {
		case err := <-runErr:
			panic(err)
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}

	cancel()

	fmt.Println("shutdown state:", <-runErr)
}

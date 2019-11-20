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

package state

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joeycumines/go-bigbuff"
	"sync"
	"time"
)

func ExampleRun_counter() {
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
		runErr <- Run(
			ctx,
			InitTicker,
			UpdateTicker,
			ViewTicker,
			buffer,
			consumer,
		)
	}()

	if err := buffer.Put(
		ctx,
		TickerHorizontalCounter(IncrementCounter{}),
		TickerHorizontalCounter(IncrementCounter{}),
		TickerVerticalCounter(DecrementCounter{}),
		TickerHorizontalCounter(DecrementCounter{}),
		SetHorizontalMultiTicker(1),
		SetVerticalMultiTicker(-1),
		TickTicker{},
		TickTicker{},
		TickTicker{},
		SetHorizontalMultiTicker(-10),
		SetVerticalMultiTicker(10),
		TickTicker{},
		TickTicker{},
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

	if err := <-runErr; err != context.Canceled {
		panic(err)
	}

	//output:
	//INIT TICKER
	//INIT COUNTER
	//INIT COUNTER
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-1 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): 1
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:0}
	//UPDATE TICKER (state.SetVerticalMultiTicker): -1
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): -10
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:-1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): 10
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:3 VerticalCounter:-2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:3 VerticalCounter:-3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:4 VerticalCounter:-3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:4 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:3 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-1 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-2 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-3 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-4 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-5 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:5 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-7 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-8 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-11 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-12 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-13 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-14 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:7 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:8 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:10 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:11 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:12 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:13 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:14 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:15 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:-10 VerticalMulti:10}
}

func ExampleBatch_counter() {
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
			TickerHorizontalCounter(IncrementCounter{}),
			TickerHorizontalCounter(IncrementCounter{}),
			TickerVerticalCounter(DecrementCounter{}),
			TickerHorizontalCounter(DecrementCounter{}),
			SetHorizontalMultiTicker(1),
			SetVerticalMultiTicker(-1),
			TickTicker{},
			TickTicker{},
			TickTicker{},
			SetHorizontalMultiTicker(-10),
			SetVerticalMultiTicker(10),
			TickTicker{},
			TickTicker{},
		}

		time.Sleep(time.Second)

		fetcherMessages <- []interface{}{
			SetHorizontalMultiTicker(1),
			SetVerticalMultiTicker(-1),
		}
		fetcherMessages <- []interface{}{
			TickTicker{},
			TickTicker{},
		}

		time.Sleep(time.Second)

		fetcherMessages <- []interface{}{
			SetHorizontalMultiTicker(-1),
			SetVerticalMultiTicker(1),
		}
		fetcherMessages <- []interface{}{
			TickTicker{},
			TickTicker{},
		}
		fetcherMessages <- []interface{}{}

		time.Sleep(time.Second)

		fetcherMessages <- []interface{}{}
		fetcherMessages <- []interface{}{
			TickerHorizontalCounter(DecrementCounter{}),
			TickerHorizontalCounter(DecrementCounter{}),
			TickerVerticalCounter(IncrementCounter{}),
			TickerHorizontalCounter(IncrementCounter{}),
			SetHorizontalMultiTicker(-1),
			SetVerticalMultiTicker(1),
			TickTicker{},
			TickTicker{},
			TickTicker{},
			SetHorizontalMultiTicker(10),
			SetVerticalMultiTicker(-10),
			TickTicker{},
			TickTicker{},
		}
	}()

	err = Batch(
		ctx,
		InitTicker,
		UpdateTicker,
		ViewTicker,
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

	//output:
	//INIT TICKER
	//INIT COUNTER
	//INIT COUNTER
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-1 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): 1
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:0}
	//UPDATE TICKER (state.SetVerticalMultiTicker): -1
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): -10
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:-1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): 10
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:3 VerticalCounter:-2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:3 VerticalCounter:-3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:4 VerticalCounter:-3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:4 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:3 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:2 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-1 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-2 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-3 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-4 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-5 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:5 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-7 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-8 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-11 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-12 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-13 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-14 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:7 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:8 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:10 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:11 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:12 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:13 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:14 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:15 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): 1
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:1 VerticalMulti:10}
	//UPDATE TICKER (state.SetVerticalMultiTicker): -1
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:16 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:15 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-14 VerticalCounter:15 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-14 VerticalCounter:14 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): -1
	//{HorizontalCounter:-14 VerticalCounter:14 HorizontalMulti:-1 VerticalMulti:-1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): 1
	//{HorizontalCounter:-14 VerticalCounter:14 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-14 VerticalCounter:14 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-14 VerticalCounter:14 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:14 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:15 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:15 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:16 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-17 VerticalCounter:16 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-18 VerticalCounter:16 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-18 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): -1
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): 1
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:-1 VerticalMulti:1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): 10
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:10 VerticalMulti:1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): -10
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-17 VerticalCounter:17 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-18 VerticalCounter:17 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-18 VerticalCounter:18 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:18 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:19 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-20 VerticalCounter:19 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-20 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-18 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-17 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-14 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-13 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-12 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-11 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:20 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:19 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:18 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:17 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:16 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:15 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:14 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:13 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:12 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:11 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-8 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-7 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-5 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-4 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-3 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-2 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-1 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:10 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:9 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:8 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:7 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:6 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:5 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:4 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:3 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:2 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:1 HorizontalMulti:10 VerticalMulti:-10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:10 VerticalMulti:-10}
}

func ExampleAggregate_counter() {
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
				TickTicker{},
				TickTicker{},

				// these always adjust by the same amounts
				SetHorizontalMultiTicker(1),
				SetVerticalMultiTicker(-1),
				TickTicker{},
				SetHorizontalMultiTicker(-10),
				SetVerticalMultiTicker(10),
				TickTicker{},
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
		var ticker Ticker
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

	err = Aggregate(
		ctx,
		InitTicker,
		UpdateTicker,
		ViewTicker,
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

	err = Aggregate(
		ctx,
		InitTicker,
		UpdateTicker,
		ViewTicker,
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

	//output:
	//Running for the first time for 'model_key'...
	//INIT TICKER
	//INIT COUNTER
	//INIT COUNTER
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:0 VerticalMulti:0}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): 1
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:1 VerticalMulti:0}
	//UPDATE TICKER (state.SetVerticalMultiTicker): -1
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): -10
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:-1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): 10
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:0 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:0 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-1 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-2 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-3 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-4 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-5 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-6 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-7 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-8 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:-1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:0 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:1 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:2 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:3 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:4 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:5 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:6 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:7 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:8 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//
	//MODEL:{"HorizontalCounter":-9,"VerticalCounter":9,"HorizontalMulti":-10,"VerticalMulti":10}
	//
	//Running again against 'model_key'...
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): 1
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:1 VerticalMulti:10}
	//UPDATE TICKER (state.SetVerticalMultiTicker): -1
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:1 VerticalMulti:-1}
	//UPDATE TICKER (state.SetHorizontalMultiTicker): -10
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:-1}
	//UPDATE TICKER (state.SetVerticalMultiTicker): 10
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.TickTicker): {}
	//{HorizontalCounter:-9 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-10 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-11 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-12 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-13 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-14 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-15 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-16 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-17 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-18 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:9 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:10 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:11 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:12 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:13 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:14 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:15 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:16 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:17 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:18 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-19 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-20 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-21 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-22 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-23 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-24 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-25 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-26 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-27 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-28 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:19 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:20 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:21 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:22 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:23 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:24 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:25 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:26 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:27 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:29 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-28 VerticalCounter:29 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-28 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-29 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-30 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-31 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-32 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-33 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-34 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-35 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-36 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-37 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerHorizontalCounter): {Message:{}}
	//UPDATE COUNTER (state.DecrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:28 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:29 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:30 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:31 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:32 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:33 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:34 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:35 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:36 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:37 HorizontalMulti:-10 VerticalMulti:10}
	//UPDATE TICKER (state.tickerVerticalCounter): {Message:{}}
	//UPDATE COUNTER (state.IncrementCounter): {}
	//{HorizontalCounter:-38 VerticalCounter:38 HorizontalMulti:-10 VerticalMulti:10}
	//
	//MODEL:{"HorizontalCounter":-38,"VerticalCounter":38,"HorizontalMulti":-10,"VerticalMulti":10}
}

type (
	Ticker struct {
		HorizontalCounter Counter
		VerticalCounter   Counter
		HorizontalMulti   int
		VerticalMulti     int
	}

	TickTicker struct{}

	SetHorizontalMultiTicker int

	SetVerticalMultiTicker int

	tickerHorizontalCounter struct {
		Message interface{}
	}

	tickerVerticalCounter struct {
		Message interface{}
	}
)

func TickerHorizontalCounter(message interface{}) interface{} {
	return tickerHorizontalCounter{
		Message: message,
	}
}

func TickerVerticalCounter(message interface{}) interface{} {
	return tickerVerticalCounter{
		Message: message,
	}
}

func InitTicker() (
	interface{},
	[]func() (message interface{}),
) {
	fmt.Println("INIT TICKER")
	model, command := Ticker{}, make([]func() (message interface{}), 0)

	// init horizontal counter
	m, c := InitCounter()
	model.HorizontalCounter = m.(Counter)
	command = append(command, c...)

	// init vertical counter
	m, c = InitCounter()
	model.VerticalCounter = m.(Counter)
	command = append(command, c...)

	return model, command
}

func UpdateTicker(
	message interface{},
) (
	func(
		currentModel interface{},
	) (
		updatedModel interface{},
		command []func() (message interface{}),
	),
) {
	fmt.Printf("UPDATE TICKER (%T): %+v\n", message, message)
	switch msg := message.(type) {
	case TickTicker:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Ticker)

			if model.HorizontalMulti > 0 {
				for x := 0; x < model.HorizontalMulti; x++ {
					command = append(command, Tag(NewCommand(IncrementCounter{}), TickerHorizontalCounter)...)
				}
			} else {
				for x := 0; x < model.HorizontalMulti*-1; x++ {
					command = append(command, Tag(NewCommand(DecrementCounter{}), TickerHorizontalCounter)...)
				}
			}

			if model.VerticalMulti > 0 {
				for x := 0; x < model.VerticalMulti; x++ {
					command = append(command, Tag(NewCommand(IncrementCounter{}), TickerVerticalCounter)...)
				}
			} else {
				for x := 0; x < model.VerticalMulti*-1; x++ {
					command = append(command, Tag(NewCommand(DecrementCounter{}), TickerVerticalCounter)...)
				}
			}

			updatedModel = model

			return
		}

	case SetHorizontalMultiTicker:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Ticker)

			model.HorizontalMulti = int(msg)

			updatedModel = model

			return
		}

	case SetVerticalMultiTicker:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Ticker)

			model.VerticalMulti = int(msg)

			updatedModel = model

			return
		}

	case tickerHorizontalCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Ticker)

			if transform := UpdateCounter(msg.Message); transform != nil {
				m, c := transform(model.HorizontalCounter)
				model.HorizontalCounter = m.(Counter)
				command = append(command, Tag(c, TickerHorizontalCounter)...)
			}

			updatedModel = model

			return
		}

	case tickerVerticalCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Ticker)

			if transform := UpdateCounter(msg.Message); transform != nil {
				m, c := transform(model.VerticalCounter)
				model.VerticalCounter = m.(Counter)
				command = append(command, Tag(c, TickerVerticalCounter)...)
			}

			updatedModel = model

			return
		}

	default:
		return nil
	}
}

func ViewTicker(model interface{}) {
	fmt.Printf("%+v\n", model.(Ticker))
}

type (
	Counter int

	IncrementCounter struct{}

	DecrementCounter struct{}
)

func InitCounter() (
	model interface{},
	command []func() (message interface{}),
) {
	fmt.Println("INIT COUNTER")
	return Counter(0), nil
}

func UpdateCounter(
	message interface{},
) (
	func(
		currentModel interface{},
	) (
		updatedModel interface{},
		command []func() (message interface{}),
	),
) {
	fmt.Printf("UPDATE COUNTER (%T): %+v\n", message, message)
	switch message.(type) {
	case IncrementCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Counter)
			model++
			return model, nil
		}

	case DecrementCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(Counter)
			model--
			return model, nil
		}

	default:
		return nil
	}
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

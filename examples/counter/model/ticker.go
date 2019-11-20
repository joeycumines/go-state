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

package model

import (
	"fmt"
	"github.com/joeycumines/go-state"
)

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
					command = append(command, state.Tag(state.NewCommand(IncrementCounter{}), TickerHorizontalCounter)...)
				}
			} else {
				for x := 0; x < model.HorizontalMulti*-1; x++ {
					command = append(command, state.Tag(state.NewCommand(DecrementCounter{}), TickerHorizontalCounter)...)
				}
			}

			if model.VerticalMulti > 0 {
				for x := 0; x < model.VerticalMulti; x++ {
					command = append(command, state.Tag(state.NewCommand(IncrementCounter{}), TickerVerticalCounter)...)
				}
			} else {
				for x := 0; x < model.VerticalMulti*-1; x++ {
					command = append(command, state.Tag(state.NewCommand(DecrementCounter{}), TickerVerticalCounter)...)
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
				command = append(command, state.Tag(c, TickerHorizontalCounter)...)
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
				command = append(command, state.Tag(c, TickerVerticalCounter)...)
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

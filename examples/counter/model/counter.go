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

package model

import "fmt"

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

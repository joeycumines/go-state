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

package state

import (
	"fmt"
	"testing"
	"strings"
	"github.com/go-test/deep"
)

func TestConfig_Apply_panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		s := fmt.Sprint(r)
		if s != "state.Config.Apply nil receiver" {
			t.Fatal("unexpected panic", s)
		}
	}()
	var c *Config
	c.Apply()
}

func TestConfig_Apply_nil(t *testing.T) {
	c := new(Config)
	err := c.Apply(
		nil,
		func(config *Config) error {
			config.Key = "21"
			return nil
		},
	)
	if err != nil || c.Key != "21" {
		t.Fatal(err, c)
	}
}

func TestTag_nilChildCommand(t *testing.T) {
	if Tag(nil, func(interface{}) interface{} { return nil }) != nil {
		t.Fatal("expected a nil parent command")
	}
}

func TestTag_nilTarget(t *testing.T) {
	if Tag(make([]func() (message interface{}), 0), nil) != nil {
		t.Fatal("expected a nil parent command")
	}
}

func TestTag_init(t *testing.T) {
	var T1 = func(msg interface{}) interface{} {
		if msg == nil || msg.(int) < 0 {
			return -1
		}
		return msg.(int) + 1
	}
	var T2 = func(msg interface{}) interface{} {
		if msg == nil || msg.(int) < 0 {
			return -1
		}
		return msg.(int) * 2
	}

	testCases := []struct {
		Init []func() (message interface{})
		T1   []int
		T2   []int
		T1T2 []int
	}{
		{
			Init: []func() (message interface{}){
				func() interface{} {
					return 1
				},
				func() interface{} {
					return 2
				},
				func() interface{} {
					return nil
				},
				func() interface{} {
					return 3
				},
				nil,
				func() interface{} {
					return 4
				},
			},
			T1: []int{
				2,
				3,
				-1,
				4,
				-2,
				5,
			},
			T2: []int{
				2,
				4,
				-1,
				6,
				-2,
				8,
			},
			T1T2: []int{
				4,
				6,
				-1,
				8,
				-2,
				10,
			},
		},
		{
			Init: []func() (message interface{}){
				func() interface{} {
					return 3
				},
			},
			T1: []int{
				4,
			},
			T2: []int{
				6,
			},
			T1T2: []int{
				8,
			},
		},
		{
			Init: nil,
			T1:   []int{},
			T2:   []int{},
			T1T2: []int{},
		},
	}

	cmdToInts := func(cmd []func() (message interface{})) []int {
		result := make([]int, len(cmd))

		for i := range result {
			c := cmd[i]

			if c == nil {
				result[i] = -2
				continue
			}

			if vInt, ok := c().(int); ok {
				result[i] = vInt
				continue
			}

			result[i] = -3
		}

		return result
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("case_#%d", i)

		oldInit := testCase.Init
		init := cmdToInts(testCase.Init)

		t1 := Tag(testCase.Init, T1)
		if diff := deep.Equal(oldInit, testCase.Init); diff != nil {
			t.Error(name, "unexpected init command (changed):", strings.Join(diff, "\n  "))
		}
		if diff := deep.Equal(init, cmdToInts(testCase.Init)); diff != nil {
			t.Error(name, "unexpected init values (changed):", strings.Join(diff, "\n  "))
		}
		if diff := deep.Equal(testCase.T1, cmdToInts(t1)); diff != nil {
			t.Error(name, "unexpected t1:", strings.Join(diff, "\n  "))
		}

		t2 := Tag(testCase.Init, T2)
		if diff := deep.Equal(oldInit, testCase.Init); diff != nil {
			t.Error(name, "unexpected init command (changed):", strings.Join(diff, "\n  "))
		}
		if diff := deep.Equal(init, cmdToInts(testCase.Init)); diff != nil {
			t.Error(name, "unexpected init values (changed):", strings.Join(diff, "\n  "))
		}
		if diff := deep.Equal(testCase.T2, cmdToInts(t2)); diff != nil {
			t.Error(name, "unexpected t2:", strings.Join(diff, "\n  "))
		}

		t1t2 := Tag(Tag(testCase.Init, T1), T2)
		if diff := deep.Equal(oldInit, testCase.Init); diff != nil {
			t.Error(name, "unexpected init command (changed):", strings.Join(diff, "\n  "))
		}
		if diff := deep.Equal(init, cmdToInts(testCase.Init)); diff != nil {
			t.Error(name, "unexpected init values (changed):", strings.Join(diff, "\n  "))
		}
		if diff := deep.Equal(testCase.T1T2, cmdToInts(t1t2)); diff != nil {
			t.Error(name, "unexpected t2:", strings.Join(diff, "\n  "))
		}
	}
}

func TestNewCommand(t *testing.T) {
	messages := []interface{}{
		1,
		"123",
		true,
		nil,
		struct{}{},
	}

	command := NewCommand(messages...)

	output := make([]interface{}, 0)

	for _, cmd := range command {
		output = append(output, cmd())
	}

	if diff := deep.Equal(output, messages); diff != nil {
		t.Fatal("unexpected diff", diff)
	}
}

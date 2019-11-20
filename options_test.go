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
	"errors"
	"fmt"
	"github.com/go-test/deep"
	"reflect"
	"sync"
	"testing"
)

func TestAggregateValidator(t *testing.T) {
	type TestCase struct {
		Options []Option
		Error   error
		Config  *Config
	}

	var (
		init = func() (model interface{}, command []func() (message interface{})) {
			return
		}
		update = func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		}
		view = func(model interface{}) {
		}
		producer = mockProducer{}
		consumer = mockConsumer{}
		fetcher  = func(ctx context.Context) ([]interface{}, bool, error) {
			return nil, false, nil
		}
		store    = new(mockStore)
		hydrator = func(ctx context.Context, key string, value []byte) (model interface{}, readOnlyModels []func() (key string, load func(model interface{})), err error) {
			return nil, nil, nil
		}
		dehydrator = func(ctx context.Context, key string, model interface{}) (value []byte, err error) {
			return nil, nil
		}
		key = "some_key"
	)

	testCases := []TestCase{
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: nil,
			Config: &Config{
				Init:       init,
				Update:     update,
				View:       view,
				Producer:   producer,
				Consumer:   consumer,
				Fetcher:    fetcher,
				Store:      store,
				Hydrator:   hydrator,
				Dehydrator: dehydrator,
				Key:        key,
				Replay:     func() bool { return false },
			},
		},
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				nil,
				key,
			),
			Error: errors.New("state.AggregateValidator nil dehydrator"),
		},
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				fetcher,
				store,
				nil,
				dehydrator,
				key,
			),
			Error: errors.New("state.AggregateValidator nil hydrator"),
		},
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				fetcher,
				nil,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.AggregateValidator nil store"),
		},
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				nil,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.BatchValidator nil fetcher"),
		},
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				producer,
				nil,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.RunValidator nil consumer"),
		},
		{
			Options: AggregateOptions(
				init,
				update,
				view,
				nil,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.RunValidator nil producer"),
		},
		{
			Options: AggregateOptions(
				init,
				update,
				nil,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.RunValidator nil view"),
		},
		{
			Options: AggregateOptions(
				init,
				nil,
				view,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.RunValidator nil update"),
		},
		{
			Options: AggregateOptions(
				nil,
				update,
				view,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			Error: errors.New("state.RunValidator nil init"),
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestAggregateValidator_#%d", i+1)

		var config Config

		err := config.Apply(testCase.Options...)

		if err != nil {
			t.Fatal(name, "shouldnt have errored on option:", err)
		}

		err = AggregateValidator(&config)

		if (err == nil) != (testCase.Error == nil) || (err != nil && err.Error() != testCase.Error.Error()) {
			t.Error(name, "unexpected error", err)
		}

		if testCase.Config != nil {
			if diff := deep.Equal(config, *testCase.Config); diff != nil {
				t.Error(name, "unexpected diff", diff)
			}
		}
	}
}

func TestOptionConfig(t *testing.T) {
	var (
		init = func() (model interface{}, command []func() (message interface{})) {
			return
		}
		update = func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		}
		view = func(model interface{}) {
		}
		producer = mockProducer{}
		consumer = mockConsumer{}
		fetcher  = func(ctx context.Context) ([]interface{}, bool, error) {
			return nil, false, nil
		}
		store    = new(mockStore)
		hydrator = func(ctx context.Context, key string, value []byte) (model interface{}, readOnlyModels []func() (key string, load func(model interface{})), err error) {
			return nil, nil, nil
		}
		dehydrator = func(ctx context.Context, key string, model interface{}) (value []byte, err error) {
			return nil, nil
		}
		key = "some_key"
	)

	expected := Config{
		Init:       init,
		Update:     update,
		View:       view,
		Producer:   producer,
		Consumer:   consumer,
		Fetcher:    fetcher,
		Store:      store,
		Hydrator:   hydrator,
		Dehydrator: dehydrator,
		Key:        key,
		Replay:     func() bool { return true },
	}

	testCases := [][]Option{
		{
			OptionConfig(expected),
		},
		append(
			AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			),
			OptionReplay(func() bool { return true }),
		),
		append(
			[]Option{OptionReplay(func() bool { return true })},
			AggregateOptions(
				init,
				update,
				view,
				producer,
				consumer,
				fetcher,
				store,
				hydrator,
				dehydrator,
				key,
			)...,
		),
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestOptionConfig_#%d", i+1)

		var config Config

		err := config.Apply(testCase...)

		if err != nil {
			t.Fatal(name, "shouldnt have errored on option:", err)
		}

		if diff := deep.Equal(config, expected); diff != nil {
			t.Error(name, "unexpected diff", diff)
		}

		if reflect.ValueOf(config.Init).Pointer() != reflect.ValueOf(expected.Init).Pointer() ||
			reflect.ValueOf(config.View).Pointer() != reflect.ValueOf(expected.View).Pointer() ||
			reflect.ValueOf(config.Update).Pointer() != reflect.ValueOf(expected.Update).Pointer() ||
			reflect.ValueOf(config.Fetcher).Pointer() != reflect.ValueOf(expected.Fetcher).Pointer() ||
			reflect.ValueOf(config.Hydrator).Pointer() != reflect.ValueOf(expected.Hydrator).Pointer() ||
			reflect.ValueOf(config.Dehydrator).Pointer() != reflect.ValueOf(expected.Dehydrator).Pointer() {
			t.Error(name, "unexpected function(s)")
		}
	}
}

type mockStore sync.Map

func (s *mockStore) loadString(key string) string {
	m := (*sync.Map)(s)
	v, _ := m.Load(key)
	b, _ := v.([]byte)
	return string(b)
}

func (s *mockStore) Load(ctx context.Context, key string) ([]byte, bool, error) {
	m := (*sync.Map)(s)
	v, ok := m.Load(key)
	b, _ := v.([]byte)
	return b, ok, nil
}

func (s *mockStore) Store(ctx context.Context, key string, value []byte) error {
	m := (*sync.Map)(s)
	m.Store(key, value)
	return nil
}

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
	"testing"
	"context"
	"time"
	"errors"
	"github.com/go-test/deep"
	"fmt"
	"github.com/joeycumines/go-bigbuff"
)

func TestBatch_nilCtx(t *testing.T) {
	err := Batch(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	if err == nil || err.Error() != "state.Batch nil ctx" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_nilUpdate(t *testing.T) {
	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			panic("not implemented")
		},
		nil,
		func(model interface{}) {
			panic("not implemented")
		},
		mockProducer{},
		mockConsumer{},
		func(ctx context.Context) ([]interface{}, bool, error) {
			panic("not implemented")
		},
	)
	if err == nil || err.Error() != "state.Batch config error: nil update" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_nilFetcher(t *testing.T) {
	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			panic("not implemented")
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			panic("not implemented")
		},
		func(model interface{}) {
			panic("not implemented")
		},
		mockProducer{},
		mockConsumer{},
		nil,
	)
	if err == nil || err.Error() != "state.Batch config error: nil fetcher" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_fetcherPanic(t *testing.T) {
	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				return nil, nil
			},
			commit: func() error {
				return nil
			},
			rollback: func() error {
				return nil
			},
		},
		func(ctx context.Context) ([]interface{}, bool, error) {
			panic("some_panic")
		},
	)
	if err == nil || err.Error() != "state.Batch run error: state.Run context error: context canceled | state.Batch worker error: recovered from panic (string): some_panic" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_runError(t *testing.T) {
	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				return nil, nil
			},
			commit: func() error {
				time.Sleep(time.Millisecond * 100)
				return errors.New("some_error")
			},
			rollback: func() error {
				return nil
			},
		},
		func(ctx context.Context) ([]interface{}, bool, error) {
			time.Sleep(time.Millisecond * 50)
			return nil, false, nil
		},
	)
	if err == nil || err.Error() != "state.Batch run error: state.Run commit error: some_error" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_runErrorCancelsBatch(t *testing.T) {
	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Millisecond * 100)
				return nil, errors.New("some_error")
			},
			commit: func() error {
				return nil
			},
			rollback: func() error {
				return nil
			},
		},
		func(ctx context.Context) ([]interface{}, bool, error) {
			time.Sleep(time.Millisecond * 50)
			return nil, true, nil
		},
	)
	if err == nil || err.Error() != "state.Batch run error: state.Run consumer error: some_error | state.Batch worker error: context error: context canceled" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_batchErrorCancelsRun(t *testing.T) {
	var x int

	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Millisecond * 5)
				return nil, nil
			},
			commit: func() error {
				return nil
			},
			rollback: func() error {
				return nil
			},
		},
		func(ctx context.Context) ([]interface{}, bool, error) {
			x++
			if x > 10 {
				return nil, true, errors.New("some_error")
			}
			time.Sleep(time.Millisecond * 10)
			return nil, true, nil
		},
	)
	if err == nil || err.Error() != "state.Batch run error: state.Run context error: context canceled | state.Batch worker error: fetcher error: some_error" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_batchErrorCausedByPut(t *testing.T) {
	var x int

	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				if len(values) == 1 && values[0] == "causes fatal error" {
					return errors.New("fatal_error")
				}
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Millisecond * 5)
				return nil, nil
			},
			commit: func() error {
				return nil
			},
			rollback: func() error {
				return nil
			},
		},
		func(ctx context.Context) ([]interface{}, bool, error) {
			x++
			if x == 11 {
				return []interface{}{"causes fatal error"}, false, nil
			}
			if x > 11 {
				panic("should have never happened")
			}
			time.Sleep(time.Millisecond * 10)
			return []interface{}{1, 2}, true, nil
		},
	)
	if err == nil || err.Error() != "state.Batch run error: state.Run context error: context canceled | state.Batch worker error: producer error: fatal_error" {
		t.Fatal("unexpected err", err)
	}
}

func TestBatch_batchEndError(t *testing.T) {
	var msgs []interface{}

	err := Batch(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				if len(values) == 1 {
					if _, ok := values[0].(batchEnd); ok {
						return errors.New("some_error")
					}
				}
				msgs = append(msgs, values...)
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Millisecond * 5)
				return nil, nil
			},
			commit: func() error {
				return nil
			},
			rollback: func() error {
				return nil
			},
		},
		func(ctx context.Context) ([]interface{}, bool, error) {
			return []interface{}{1, 2}, false, nil
		},
	)

	if err == nil || err.Error() != "state.Batch run error: state.Run context error: context canceled | state.Batch worker error: end error: some_error" {
		t.Fatal("unexpected err", err)
	}

	if diff := deep.Equal(msgs, []interface{}{1, 2}); diff != nil {
		t.Fatal("unexpected diff")
	}
}

func TestBatchConsumer_Get_stopped(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		s := fmt.Sprint(r)
		if s != "state.batchConsumer.Get stopped consuming" {
			t.Fatal("unexpected panic", s)
		}
	}()
	c := batchConsumer{
		stopped: true,
	}
	c.Get(nil)
}

func TestBatchConsumer_Get_commitError(t *testing.T) {
	expectedCtx := context.WithValue(context.Background(), 1, 2)
	rollback := make(chan struct{}, 1)
	c := batchConsumer{
		Consumer: mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				if ctx != expectedCtx {
					t.Error("unexpected ctx", ctx)
				}
				return batchEnd{}, nil
			},
			commit: func() error {
				return errors.New("some_error")
			},
			rollback: func() error {
				rollback <- struct{}{}
				return nil
			},
		},
	}
	v, err := c.Get(expectedCtx)
	if v != nil {
		t.Error("unexpected v", v)
	}
	if err == nil || err.Error() != "state.batchConsumer.Get end commit error: some_error" {
		t.Error("unexpected err", err)
	}
	<-rollback
}

func TestBatchConsumer_Get_putError(t *testing.T) {
	expectedCtx := context.WithValue(context.Background(), 1, 2)
	c := batchConsumer{
		Consumer: mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				if ctx != expectedCtx {
					t.Error("unexpected ctx", ctx)
				}
				return batchEnd{}, nil
			},
			commit: func() error {
				return nil
			},
		},
		producer: mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				if ctx != expectedCtx {
					t.Error("unexpected ctx", ctx)
				}
				if len(values) != 1 {
					t.Fatal("unexpected values", values)
				}
				if _, ok := values[0].(batchEnd); !ok {
					t.Fatal("unexpected values", values)
				}
				return errors.New("some_error")
			},
		},
	}
	v, err := c.Get(expectedCtx)
	if v != nil {
		t.Error("unexpected v", v)
	}
	if err == nil || err.Error() != "state.batchConsumer.Get end put error: some_error" {
		t.Error("unexpected err", err)
	}
}

func TestBatchConsumer_Get_twice(t *testing.T) {
	var (
		get    = make(chan struct{})
		commit = make(chan struct{})
		put    = make(chan struct{})
		done   = make(chan struct{})
	)

	go func() {
		defer close(done)
		<-get
		<-commit
		<-put
		<-get
		<-commit
	}()

	c := batchConsumer{
		Consumer: mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				get <- struct{}{}
				return batchEnd{}, nil
			},
			commit: func() error {
				commit <- struct{}{}
				return nil
			},
		},
		producer: mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				put <- struct{}{}
				return nil
			},
		},
	}

	v, err := c.Get(nil)

	if v != nil {
		t.Error("unexpected v", v)
	}

	if err == nil || err.Error() != "state.batchConsumer.Get batch stopped" {
		t.Error("unexpected err", err)
	}

	<-done
}

func TestBatch_counter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buffer := new(bigbuff.Buffer)

	buffer.SetCleanerConfig(bigbuff.CleanerConfig{
		Cleaner: func(size int, offsets []int) int {
			return 0
		},
	})

	var ticker exampleTicker

	func() {
		defer buffer.Close()

		consumer, err := buffer.NewConsumer()

		if err != nil {
			t.Fatal(err)
		}

		defer consumer.Close()

		var (
			fetcherMessages = make(chan []interface{})
			fetcherDone     = make(chan struct{})
		)

		go func() {
			defer close(fetcherDone)

			fetcherMessages <- []interface{}{
				exampleSetHorizontalMultiTicker(3),
				exampleSetVerticalMultiTicker(-3),
				exampleTickTicker{},
				exampleSetHorizontalMultiTicker(-10),
				exampleSetVerticalMultiTicker(10),
				exampleTickTicker{},
			}

			time.Sleep(time.Millisecond * 100)

			fetcherMessages <- []interface{}{
				exampleSetHorizontalMultiTicker(91239),
				exampleSetVerticalMultiTicker(23111),
			}
		}()

		exampleTargetTickerCalled = false

		err = Batch(
			ctx,
			exampleInitTicker,
			exampleUpdateTicker,
			func(model interface{}) {
				ticker = model.(exampleTicker)
			},
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
			t.Fatal(err)
		}

		if !exampleTargetTickerCalled {
			t.Error("expected target commands to be called")
		}
	}()

	if diff := deep.Equal(ticker, exampleTicker{
		HorizontalCounter: -7,
		VerticalCounter:   7,
		HorizontalMulti:   91239,
		VerticalMulti:     23111,
	}); diff != nil {
		t.Error("unexpected ticker", diff)
	}

	msgs := buffer.Slice()

	if diff := deep.Equal(
		msgs,
		[]interface{}{
			exampleSetHorizontalMultiTicker(3),
			exampleSetVerticalMultiTicker(-3),
			exampleTickTicker{},
			exampleSetHorizontalMultiTicker(-10),
			exampleSetVerticalMultiTicker(10),
			exampleTickTicker{},

			exampleTickerHorizontalCounter{exampleIncrementCounter{}},
			exampleTickerHorizontalCounter{exampleIncrementCounter{}},
			exampleTickerHorizontalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleDecrementCounter{}},
			exampleTickerVerticalCounter{exampleDecrementCounter{}},
			exampleTickerVerticalCounter{exampleDecrementCounter{}},

			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},

			exampleSetHorizontalMultiTicker(91239),
			exampleSetVerticalMultiTicker(23111),

			batchEnd{},
			batchEnd{},
		},
	); diff != nil {
		t.Error("unexpected msgs", diff)
	}

	// replay until just before the end, to prove it's no different
	buffer = new(bigbuff.Buffer)

	buffer.SetCleanerConfig(bigbuff.CleanerConfig{
		Cleaner: func(size int, offsets []int) int {
			return 0
		},
	})

	func() {
		defer buffer.Close()

		consumer, err := buffer.NewConsumer()

		if err != nil {
			t.Fatal(err)
		}

		defer consumer.Close()

		var (
			fetcherMessages = make(chan []interface{})
			fetcherDone     = make(chan struct{})
		)

		go func() {
			defer close(fetcherDone)

			// this time we send each message individually, no difference in outcome
			for _, msg := range []interface{}{
				exampleSetHorizontalMultiTicker(3),
				exampleSetVerticalMultiTicker(-3),
				exampleTickTicker{},
				exampleSetHorizontalMultiTicker(-10),
				exampleSetVerticalMultiTicker(10),
				exampleTickTicker{},

				exampleTickerHorizontalCounter{exampleIncrementCounter{}},
				exampleTickerHorizontalCounter{exampleIncrementCounter{}},
				exampleTickerHorizontalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleDecrementCounter{}},
				exampleTickerVerticalCounter{exampleDecrementCounter{}},
				exampleTickerVerticalCounter{exampleDecrementCounter{}},

				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerHorizontalCounter{exampleDecrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
				exampleTickerVerticalCounter{exampleIncrementCounter{}},
			} {
				fetcherMessages <- []interface{}{msg}
			}
		}()

		exampleTargetTickerCalled = false

		err = BatchWithOptions(
			ctx,
			append(
				BatchOptions(
					exampleInitTicker,
					exampleUpdateTicker,
					func(model interface{}) {
						ticker = model.(exampleTicker)
					},
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
				),
				OptionReplay(true),
			)...
		)

		if err != nil {
			t.Fatal(err)
		}

		if exampleTargetTickerCalled {
			t.Error("expected target commands NOT to be called")
		}
	}()

	if diff := deep.Equal(ticker, exampleTicker{
		HorizontalCounter: -7,
		VerticalCounter:   6,
		HorizontalMulti:   -10,
		VerticalMulti:     10,
	}); diff != nil {
		t.Error("unexpected ticker", diff)
	}

	msgs = buffer.Slice()

	if diff := deep.Equal(
		msgs,
		[]interface{}{
			exampleSetHorizontalMultiTicker(3),
			exampleSetVerticalMultiTicker(-3),
			exampleTickTicker{},
			exampleSetHorizontalMultiTicker(-10),
			exampleSetVerticalMultiTicker(10),
			exampleTickTicker{},

			exampleTickerHorizontalCounter{exampleIncrementCounter{}},
			exampleTickerHorizontalCounter{exampleIncrementCounter{}},
			exampleTickerHorizontalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleDecrementCounter{}},
			exampleTickerVerticalCounter{exampleDecrementCounter{}},
			exampleTickerVerticalCounter{exampleDecrementCounter{}},

			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerHorizontalCounter{exampleDecrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},
			exampleTickerVerticalCounter{exampleIncrementCounter{}},

			batchEnd{},
			batchEnd{},
		},
	); diff != nil {
		t.Error("unexpected msgs", diff)
	}
}

type (
	exampleTicker struct {
		HorizontalCounter exampleCounter
		VerticalCounter   exampleCounter
		HorizontalMulti   int
		VerticalMulti     int
	}

	exampleTickTicker struct{}

	exampleSetHorizontalMultiTicker int

	exampleSetVerticalMultiTicker int

	exampleTickerHorizontalCounter struct {
		Message interface{}
	}

	exampleTickerVerticalCounter struct {
		Message interface{}
	}
)

var (
	exampleTargetTickerCalled = false
)

func exampleTargetTickerHorizontalCounter(message interface{}) interface{} {
	exampleTargetTickerCalled = true
	return exampleTickerHorizontalCounter{
		Message: message,
	}
}

func exampleTargetTickerVerticalCounter(message interface{}) interface{} {
	exampleTargetTickerCalled = true
	return exampleTickerVerticalCounter{
		Message: message,
	}
}

func exampleInitTicker() (
	interface{},
	[]func() (message interface{}),
) {
	//fmt.Println("INIT TICKER")
	model, command := exampleTicker{}, make([]func() (message interface{}), 0)

	// init horizontal counter
	m, c := exampleInitCounter()
	model.HorizontalCounter = m.(exampleCounter)
	command = append(command, c...)

	// init vertical counter
	m, c = exampleInitCounter()
	model.VerticalCounter = m.(exampleCounter)
	command = append(command, c...)

	return model, command
}

func exampleUpdateTicker(
	message interface{},
) (
	func(
		currentModel interface{},
	) (
		updatedModel interface{},
		command []func() (message interface{}),
	),
) {
	//fmt.Printf("UPDATE TICKER (%T): %+v\n", message, message)
	switch msg := message.(type) {
	case exampleTickTicker:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleTicker)

			if model.HorizontalMulti > 0 {
				for x := 0; x < model.HorizontalMulti; x++ {
					command = append(command, Tag(NewCommand(exampleIncrementCounter{}), exampleTargetTickerHorizontalCounter)...)
				}
			} else {
				for x := 0; x < model.HorizontalMulti * -1; x++ {
					command = append(command, Tag(NewCommand(exampleDecrementCounter{}), exampleTargetTickerHorizontalCounter)...)
				}
			}

			if model.VerticalMulti > 0 {
				for x := 0; x < model.VerticalMulti; x++ {
					command = append(command, Tag(NewCommand(exampleIncrementCounter{}), exampleTargetTickerVerticalCounter)...)
				}
			} else {
				for x := 0; x < model.VerticalMulti * -1; x++ {
					command = append(command, Tag(NewCommand(exampleDecrementCounter{}), exampleTargetTickerVerticalCounter)...)
				}
			}

			updatedModel = model

			return
		}

	case exampleSetHorizontalMultiTicker:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleTicker)

			model.HorizontalMulti = int(msg)

			updatedModel = model

			return
		}

	case exampleSetVerticalMultiTicker:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleTicker)

			model.VerticalMulti = int(msg)

			updatedModel = model

			return
		}

	case exampleTickerHorizontalCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleTicker)

			if transform := exampleUpdateCounter(msg.Message); transform != nil {
				m, c := transform(model.HorizontalCounter)
				model.HorizontalCounter = m.(exampleCounter)
				command = append(command, Tag(c, exampleTargetTickerHorizontalCounter)...)
			}

			updatedModel = model

			return
		}

	case exampleTickerVerticalCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleTicker)

			if transform := exampleUpdateCounter(msg.Message); transform != nil {
				m, c := transform(model.VerticalCounter)
				model.VerticalCounter = m.(exampleCounter)
				command = append(command, Tag(c, exampleTargetTickerVerticalCounter)...)
			}

			updatedModel = model

			return
		}

	default:
		return nil
	}
}

type (
	exampleCounter int

	exampleIncrementCounter struct{}

	exampleDecrementCounter struct{}
)

func exampleInitCounter() (
	model interface{},
	command []func() (message interface{}),
) {
	//fmt.Println("INIT COUNTER")
	return exampleCounter(0), nil
}

func exampleUpdateCounter(
	message interface{},
) (
	func(
		currentModel interface{},
	) (
		updatedModel interface{},
		command []func() (message interface{}),
	),
) {
	//fmt.Printf("UPDATE COUNTER (%T): %+v\n", message, message)
	switch message.(type) {
	case exampleIncrementCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleCounter)
			model++
			return model, nil
		}

	case exampleDecrementCounter:
		return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			model := currentModel.(exampleCounter)
			model--
			return model, nil
		}

	default:
		return nil
	}
}

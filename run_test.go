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
	"fmt"
	"errors"
	"github.com/go-test/deep"
)

type mockProducer struct {
	put func(ctx context.Context, values ... interface{}) error
}

func (m mockProducer) Put(ctx context.Context, values ... interface{}) error {
	if m.put != nil {
		return m.put(ctx, values...)
	}
	panic("implement me")
}

type mockConsumer struct {
	get      func(ctx context.Context) (interface{}, error)
	commit   func() error
	rollback func() error
}

func (m mockConsumer) Get(ctx context.Context) (interface{}, error) {
	if m.get != nil {
		return m.get(ctx)
	}
	panic("implement me")
}

func (m mockConsumer) Commit() error {
	if m.commit != nil {
		return m.commit()
	}
	panic("implement me")
}

func (m mockConsumer) Rollback() error {
	if m.rollback != nil {
		return m.rollback()
	}
	panic("implement me")
}

func TestRun_nilCtx(t *testing.T) {
	err := Run(
		nil,
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
			return
		},
		mockProducer{},
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run nil ctx" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_nilInit(t *testing.T) {
	err := Run(
		context.Background(),
		nil,
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
			return
		},
		mockProducer{},
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run config error: nil init" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_nilUpdate(t *testing.T) {
	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		nil,
		func(model interface{}) {
			return
		},
		mockProducer{},
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run config error: nil update" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_nilView(t *testing.T) {
	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		nil,
		mockProducer{},
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run config error: nil view" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_nilProducer(t *testing.T) {
	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
			return
		},
		nil,
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run config error: nil producer" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_nilConsumer(t *testing.T) {
	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
			return
		},
		mockProducer{},
		nil,
	)

	if err == nil || err.Error() != "state.Run config error: nil consumer" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_contextError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Run(
		ctx,
		func() (model interface{}, command []func() (message interface{})) {
			panic("unexpected call")
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			panic("unexpected call")
		},
		func(model interface{}) {
			panic("unexpected call")
		},
		mockProducer{},
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run context error: context canceled" {
		t.Fatal("unexpected error", err)
	}
}

// ensures that init is called first, and no rollback is called on init
func TestRun_initPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected a panic")
		}
		s := fmt.Sprint(r)
		if s != "init panic" {
			t.Fatal("unexpected panic", s)
		}
	}()

	Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			panic("init panic")
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			return nil
		},
		func(model interface{}) {
			return
		},
		mockProducer{},
		mockConsumer{},
	)
}

func TestRun_commitErrorDataFlow(t *testing.T) {
	init := make(chan struct{})
	update := make(chan struct{})
	transform := make(chan struct{})
	view := make(chan struct{})
	cmd := make(chan struct{})
	put := make(chan struct{})
	get := make(chan struct{})
	rollback := make(chan struct{})
	commit := make(chan struct{})

	done := make(chan struct{})

	go func() {
		defer close(done)
		init <- struct{}{}
		view <- struct{}{}
		cmd <- struct{}{}
		put <- struct{}{}
		get <- struct{}{}
		update <- struct{}{}
		transform <- struct{}{}
		view <- struct{}{}
		cmd <- struct{}{}
		put <- struct{}{}
		commit <- struct{}{}
		rollback <- struct{}{}
	}()

	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			<-init
			command = append(
				command,
				nil,
				func() (message interface{}) {
					<-cmd
					return 12
				},
				func() (message interface{}) {
					return nil
				},
				nil,
			)
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			<-update
			return func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
				<-transform
				command = append(
					command,
					nil,
					func() (message interface{}) {
						return nil
					},
					func() (message interface{}) {
						<-cmd
						return 12
					},
					nil,
				)
				return
			}
		},
		func(model interface{}) {
			<-view
			return
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				if ctx == nil {
					t.Error("expected a ctx")
				}
				if diff := deep.Equal([]interface{}{12}, values); diff != nil {
					t.Error("unexpected values", diff)
				}
				<-put
				return nil
			},
		},
		mockConsumer{
			get: func(ctx context.Context) (interface{}, error) {
				if ctx == nil {
					t.Error("expected a ctx")
				}
				<-get
				return struct{}{}, nil
			},
			commit: func() error {
				<-commit
				return errors.New("some_error")
			},
			rollback: func() error {
				<-rollback
				return nil
			},
		},
	)

	if err == nil || err.Error() != "state.Run commit error: some_error" {
		t.Fatal("unexpected error", err)
	}

	<-done
}

func TestRun_getError(t *testing.T) {
	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			panic("no update expected")
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
				return nil, errors.New("some_error")
			},
		},
	)

	if err == nil || err.Error() != "state.Run consumer error: some_error" {
		t.Fatal("unexpected error", err)
	}
}

func TestRun_putError(t *testing.T) {
	err := Run(
		context.Background(),
		func() (model interface{}, command []func() (message interface{})) {
			return
		},
		func(message interface{}) func(currentModel interface{}) (updatedModel interface{}, command []func() (message interface{})) {
			panic("no update expected")
		},
		func(model interface{}) {
		},
		mockProducer{
			put: func(ctx context.Context, values ...interface{}) error {
				return errors.New("some_error")
			},
		},
		mockConsumer{},
	)

	if err == nil || err.Error() != "state.Run producer error: some_error" {
		t.Fatal("unexpected error", err)
	}
}

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

// Package state provides an elm-like implementation of an event based framework, that utilises models, views, and
// pure-function updates, to provide a testable, extensible, but simple to work with pattern, to develop reactive
// logic, particularly suited to maintaining complex event-driven states.
package state

import (
	"context"
	"errors"
)

type (
	// Init, Update, and View, these three types make up the core of all programs using this package.
	// They define the initial state (including model and commands, that may trigger some behavior), and
	// the update and view behavior that will be applied to this state.
	// For more info, I would suggest starting here: https://guide.elm-lang.org/architecture/

	// Init models a function that defines the initial state of a program, which is made up of a model, and any
	// initial command to run against that model.
	Init func() (
		model interface{},
		command []func() (message interface{}),
	)

	// Update models a function that defines how a model can be updated, the first phase delegating a given message
	// to the correct update logic, then the second phase actually performing the update, returning an updated model
	// and any command generated as a result of the update.
	Update func(
		message interface{},
	) (
		func(
			currentModel interface{},
		) (
			updatedModel interface{},
			command []func() (message interface{}),
		),
	)

	// View models a pure function that updates some external representation after each model update, note that it
	// is important to realise that after the View is called there are no guarantees on the state of model.
	View func(model interface{})

	// These interfaces are partial copies from another project of mine, github.com/joeycumines/go-bigbuff,
	// however you may use anything that implements them, note that depending on your implementation you may have
	// to solve the issue of potential deadlocks on the program loop producing a message (commands).

	// Producer models where messages are sent to be processed by the program.
	Producer interface {
		// Put will send the provided values in-order to the message buffer, or return an error.
		// It MUST NOT block in such a way that it will be possible to cause a deadlock locally.
		Put(ctx context.Context, values ... interface{}) error
	}

	// Consumer models how messages are received by the program.
	Consumer interface {
		// Get will get a message from the message buffer, at the current offset, blocking if none are available, or
		// an error if it fails.
		Get(ctx context.Context) (interface{}, error)

		// Commit will save any offset changes, and will return an error if it fails, or if the offset saved is the
		// latest.
		Commit() error

		// Rollback will undo any offset changes, and will return an error if it fails, or if the offset saved is the
		// latest.
		Rollback() error
	}

	// Store models a key-value store for aggregated models.
	Store interface {
		// Load fetches a value given a key, returning a bool which can indicate if the key existed, or an error.
		Load(
			ctx context.Context,
			key string,
		) (
			value []byte,
			ok bool,
			err error,
		)

		// Store saves a value at a key, overwriting any existing value, or returns an error.
		Store(
			ctx context.Context,
			key string,
			value []byte,
		) (
			err error,
		)
	}

	// Hydrator is used to initialise a model from a store.
	Hydrator func(
		ctx context.Context,
		key string,
		value []byte,
	) (
		model interface{},
		readOnlyModels []func() (
			key string,
			load func(model interface{}),
		),
		err error,
	)

	// Dehydrator is used to flatten a model for saving back into a store.
	Dehydrator func(
		ctx context.Context,
		key string,
		model interface{},
	) (
		value []byte,
		err error,
	)

	// Fetcher is a message fetcher, which may return 0..n messages and a bool indicating if there are any left,
	// or an error, it is important to note that if there is no error but it returns false with messages, those
	// messages WILL be sent onwards to the next stage, however they will be the last.
	// It is used for message batching, where we want to consume until a certain point.
	Fetcher func(ctx context.Context) ([]interface{}, bool, error)

	// Config is the shared configuration struct for all app modes.
	Config struct {
		// Init is the app Init func, required by Run, Batch, Aggregate.
		Init Init

		// Update is the app Update func, required by Run, Batch, Aggregate.
		Update Update

		// View is the app View func, required by Run, Batch, Aggregate.
		View View

		// Producer is where the app will produce to, required by Run, Batch, Aggregate.
		Producer Producer

		// Consumer is where the app will consume from, required by Run, Batch, Aggregate.
		Consumer Consumer

		// Replay will disable calling commands, if set, optional.
		Replay bool

		// Fetcher is the batch consumer input, required by Batch, Aggregate.
		Fetcher Fetcher

		// Store is the key-value store used to save aggregate models, required by Aggregate.
		Store Store

		// Hydrator is used to convert keyed binary data to an init model, required by Aggregate.
		Hydrator Hydrator

		// Dehydrator is used to convert a keyed model to binary data, required by Aggregate.
		Dehydrator Dehydrator

		// Key is the unique identifier for the model to run, required by Aggregate.
		Key string
	}

	// Option is used to modify config in a composeable way.
	Option func(config *Config) error
)

// Apply will pass the receiver into all options, returning the first error, or nil, note it will panic if the
// receiver is nil.
func (c *Config) Apply(opts ... Option) error {
	if c == nil {
		panic(errors.New("state.Config.Apply nil receiver"))
	}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(c); err != nil {
			return err
		}
	}
	return nil
}

// Tag applies a target func to a command, and is used to propagate the command up the chain, note either parameter
// being nil will return nil.
func Tag(
	childCommand []func() (message interface{}),
	target func(childMessage interface{}) (parentMessage interface{}),
) (
	parentCommand []func() (message interface{}),
) {
	if childCommand == nil || target == nil {
		return
	}
	parentCommand = make([]func() (message interface{}), len(childCommand))
	copy(parentCommand, childCommand)
	for i, fn := range parentCommand {
		if fn == nil {
			continue
		}
		func(fn func() interface{}) {
			parentCommand[i] = func() interface{} {
				return target(fn())
			}
		}(fn)
	}
	return
}

// NewCommand converts 0..n messages to a command.
func NewCommand(
	messages ... interface{},
) (
	command []func() (message interface{}),
) {
	result := make([]func() (message interface{}), len(messages))
	for i := range messages {
		result[i] = (commandMessage{message: messages[i]}).command
	}
	return result
}

type commandMessage struct {
	message interface{}
}

func (c commandMessage) command() interface{} {
	return c.message
}

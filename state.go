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
	"fmt"
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
)

// Run is the provided runtime logic to actually create and run a program using the pattern defined in this package.
// Note that all arguments are required.
func Run(
	ctx context.Context,
	init Init,
	update Update,
	view View,
	producer Producer,
	consumer Consumer,
) error {
	if ctx == nil {
		return errors.New("state.Run nil ctx")
	}
	if init == nil {
		return errors.New("state.Run nil init")
	}
	if update == nil {
		return errors.New("state.Run nil update")
	}
	if view == nil {
		return errors.New("state.Run nil view")
	}
	if producer == nil {
		return errors.New("state.Run nil producer")
	}
	if consumer == nil {
		return errors.New("state.Run nil consumer")
	}

	p := &program{
		ctx:      ctx,
		init:     init,
		update:   update,
		view:     view,
		producer: producer,
		consumer: consumer,
	}

	for {
		if err := p.tick(); err != nil {
			return err
		}
	}
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

// UNEXPORTED

type commandMessage struct {
	message interface{}
}

func (c commandMessage) command() interface{} {
	return c.message
}

type program struct {
	ctx         context.Context
	init        Init
	update      Update
	view        View
	producer    Producer
	consumer    Consumer
	initialised bool
	model       interface{}
}

func (p *program) nextMessage() (message interface{}, err error) {
	message, err = p.consumer.Get(p.ctx)
	if err != nil {
		message, err = nil, fmt.Errorf("state.Run consumer error: %s", err.Error())
	}
	return
}

func (p *program) updateModel() (command []func() (message interface{}), err error) {
	if !p.initialised {
		p.model, command = p.init()
		p.initialised = true
	} else {
		var message interface{}
		message, err = p.nextMessage()
		if err == nil && message != nil {
			transform := p.update(message)
			if transform != nil {
				p.model, command = transform(p.model)
			}
		}
	}
	return
}

func (p *program) tick() error {
	if err := p.ctx.Err(); err != nil {
		return fmt.Errorf("state.Run context error: %s", err.Error())
	}
	initialised := !p.initialised // are we initialising it this loop?
	command, err := p.updateModel()
	if err != nil {
		return err
	}
	rollback := !initialised // only rollback if we didn't just initialise (only if we have anything to roll back)
	defer func() {
		if rollback {
			p.consumer.Rollback()
		}
	}()
	p.view(p.model)
	var messages []interface{}
	for _, cmd := range command {
		if cmd != nil {
			if message := cmd(); message != nil {
				messages = append(messages, message)
			}
		}
	}
	if err := p.producer.Put(p.ctx, messages...); err != nil {
		return fmt.Errorf("state.Run producer error: %s", err.Error())
	}
	if !initialised { // only commit if we didn't just initialise
		if err := p.consumer.Commit(); err != nil {
			return fmt.Errorf("state.Run commit error: %s", err.Error())
		}
	}
	rollback = false
	return nil
}

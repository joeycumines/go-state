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
)

// Run is the provided runtime logic to actually create and run a program using the pattern defined in this package.
func Run(
	ctx context.Context,
	init Init,
	update Update,
	view View,
	producer Producer,
	consumer Consumer,
) error {
	return RunWithOptions(
		ctx,
		RunOptions(
			init,
			update,
			view,
			producer,
			consumer,
		)...,
	)
}

func RunWithOptions(
	ctx context.Context,
	opts ...Option,
) error {
	if ctx == nil {
		return errors.New("state.Run nil ctx")
	}

	var config Config

	if err := config.Apply(append(append(make([]Option, 0, len(opts)+1), opts...), RunValidator)...); err != nil {
		return fmt.Errorf("state.Run config error: %s", err.Error())
	}

	p := &program{
		ctx:      ctx,
		init:     config.Init,
		update:   config.Update,
		view:     config.View,
		producer: config.Producer,
		consumer: config.Consumer,
		replay:   config.Replay,
	}

	for {
		if err := p.tick(); err != nil {
			return err
		}
	}
}

type program struct {
	ctx         context.Context
	init        Init
	update      Update
	view        View
	producer    Producer
	consumer    Consumer
	initialised bool
	replay      bool
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
	didInit := !p.initialised
	canRollback := !didInit
	defer func() {
		if canRollback {
			p.consumer.Rollback()
		}
	}()
	command, err := p.updateModel()
	if err != nil {
		return err
	}
	p.view(p.model)
	var messages []interface{}
	if !p.replay {
		for _, cmd := range command {
			if cmd != nil {
				if message := cmd(); message != nil {
					messages = append(messages, message)
				}
			}
		}
	}
	if err := p.producer.Put(p.ctx, messages...); err != nil {
		return fmt.Errorf("state.Run producer error: %s", err.Error())
	}
	if !didInit {
		if err := p.consumer.Commit(); err != nil {
			return fmt.Errorf("state.Run commit error: %s", err.Error())
		}
	}
	canRollback = false
	return nil
}

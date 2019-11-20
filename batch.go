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
)

// Batch extends the behaviour of Run, processing messages from fetcher until it returns false (which still sends the
// last batch of messages) or errors, then produces a "end" message (unexported) through the producer, filtering the
// message via hooking the consumer such that it is never seen by the update function, also producing an additional
// message each time a stop message is filtered, until two are received in a row.
// This has the effect of causing the processing loop to continue until there are no more messages being propagated,
// which does rely on the stop behaviour described to be facilitated by the producer / consumer implementation, and
// is dependant on certain guarantees provided by the Run implementation within this package.
func Batch(
	ctx context.Context,
	init Init,
	update Update,
	view View,
	producer Producer,
	consumer Consumer,
	fetcher Fetcher,
) error {
	return BatchWithOptions(
		ctx,
		BatchOptions(
			init,
			update,
			view,
			producer,
			consumer,
			fetcher,
		)...,
	)
}

func BatchWithOptions(
	ctx context.Context,
	opts ...Option,
) error {
	if ctx == nil {
		return errors.New("state.Batch nil ctx")
	}

	var config Config

	if err := config.Apply(append(append(make([]Option, 0, len(opts)+1), opts...), BatchValidator)...); err != nil {
		return err
	}

	// we have dependant child goroutines, ensure they exit
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// fetches and sends, non-panic exit cases ensure this finishes
	workerOutcome := make(chan error, 1)
	go batchWorker(
		ctx,
		cancel, // cancel will be used to unblock run on error cases
		workerOutcome,
		config.Fetcher,
		config.Producer,
	)

	// batchConsumer wraps the consumer to filter batchEnd messages + unblock run on clean exit
	batchConsumer := &batchConsumer{
		Consumer: config.Consumer,
		Producer: config.Producer,
	}
	config.Consumer = batchConsumer

	runError := RunWithOptions(ctx, OptionConfig(config))
	cancel()
	workerError := <-workerOutcome

	switch runError {
	case errBatchStopped:
		return nil
	case context.Canceled:
		return workerError
	default:
		return runError
	}
}

type batchEnd struct{}

type batchConsumer struct {
	Consumer
	Producer
}

func batchWorker(
	ctx context.Context,
	cancel context.CancelFunc,
	outcome chan<- error,
	fetcher Fetcher,
	producer Producer,
) {
	var err error
	defer func() {
		if err != nil {
			cancel()
		}
		outcome <- err
	}()
	for {
		err = ctx.Err()
		if err != nil {
			return
		}
		var (
			values []interface{}
			ok     bool
		)
		values, ok, err = fetcher(ctx)
		if err != nil {
			return
		}
		err = producer.Put(ctx, values...)
		if err != nil {
			return
		}
		if !ok {
			err = producer.Put(ctx, batchEnd{})
			return
		}
	}
}

// commitOrRollback is used to ensure that a failed commit will always be followed by a rollback, even if it panics
func commitOrRollback(consumer Consumer) (err error) {
	var success bool
	defer func() {
		if !success {
			_ = consumer.Rollback()
		}
	}()
	err = consumer.Commit()
	if err == nil {
		success = true
	}
	return
}

func (b *batchConsumer) Get(ctx context.Context) (interface{}, error) {
	// consumer loop, which waits until two batchEnd messages are received in a row OR anything else is received
	for count := 1; ; count++ {
		// consume the (potentially) first batchEnd value
		if value, err := b.Consumer.Get(ctx); err != nil {
			// bail out on error case, don't pass through value just in case (won't be used anyway)
			return nil, err
		} else if _, stop := value.(batchEnd); !stop {
			// pass through on any non-stop case
			return value, nil
		}
		// we need to commit (the batchEnd) otherwise we will be breaking the consumer
		if err := commitOrRollback(b.Consumer); err != nil {
			return nil, err
		}
		// check we have received enough stop messages in a row, and if so exit without sending another
		if count > 1 {
			break
		}
		// first stop consumed, we want another, but first we must send it
		if err := b.Put(ctx, batchEnd{}); err != nil {
			return nil, err
		}
	}
	// sequential batchEnd messages consumed, mark batch as stopped, and return an error to kill Run
	return nil, errBatchStopped
}

var errBatchStopped = errors.New(`batch stopped`)

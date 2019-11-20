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

// Aggregate implements a pattern to load from and save to a key-value store (string key, with []byte data), which
// takes advantage of the run-to-end behavior of Batch. It's intended use case is server-less computing scenarios, and
// it is particularly targeted towards complicated aggregate state. To this end the Hydrator type provides it's
// "readOnlyModels" return value, which can be used to load OTHER models (handlers for which must be defined ahead of
// time) for use within 1+ model(s). Loading of external models in this manner lacks guarantees that would generally be
// present, e.g. ordering. This capability facilitates support of very large or dynamic state, or state in multiple data
// stores. The implementation also supports circular references - multiple attempts to load the same key will simply
// return the previously hydrated model. The parameter key's value will ONLY be updated on successful completion of the
// batch. Note that the hydrator will be used for ALL models.
func Aggregate(
	ctx context.Context,
	init Init,
	update Update,
	view View,
	producer Producer,
	consumer Consumer,
	fetcher Fetcher,
	store Store,
	hydrator Hydrator,
	dehydrator Dehydrator,
	key string,
) error {
	return AggregateWithOptions(
		ctx,
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
	)
}

func AggregateWithOptions(
	ctx context.Context,
	opts ...Option,
) error {
	if ctx == nil {
		return errors.New("state.Aggregate nil ctx")
	}

	var config Config

	if err := config.Apply(append(append(make([]Option, 0, len(opts)+1), opts...), AggregateValidator)...); err != nil {
		return err
	}

	// hydrate the model, overriding the init function if the key exists
	if model, err := loadModel(ctx, config.Store, config.Hydrator, config.Key); err != nil {
		return err
	} else if model != nil {
		config.Init = func() (interface{}, []func() (message interface{})) {
			return model, nil
		}
	}

	// the view is wrapped such that each view updates lastModel before calling the original view
	var (
		lastModel    interface{}
		originalView = config.View
	)
	config.View = func(model interface{}) {
		lastModel = model
		originalView(model)
	}

	// run the batch
	if err := BatchWithOptions(ctx, OptionConfig(config)); err != nil {
		return err
	}

	// dehydrate the last model, then store the new value on the key
	if value, err := config.Dehydrator(ctx, config.Key, lastModel); err != nil {
		return err
	} else if err := config.Store.Store(ctx, config.Key, value); err != nil {
		return err
	}

	return nil
}

// loadModel hydrates a model from a store, given a key, supporting circular references via temporary caching combined
// with the read only callback pattern, which allows the model to be wired up incrementally, note that if key(s)
// don't exist in the store they will be initialised as nil
func loadModel(ctx context.Context, store Store, hydrator Hydrator, key string) (interface{}, error) {
	models := make(map[string]interface{})

	if err := loadModels(ctx, store, hydrator, key, models); err != nil {
		return nil, err
	}

	return models[key], nil
}

func loadModels(ctx context.Context, store Store, hydrator Hydrator, key string, models map[string]interface{}) error {
	if value, ok, err := store.Load(ctx, key); err != nil {
		return err
	} else if !ok {
		// nil is used to represent a key without a (current) model
		models[key] = nil
		return nil
	} else if model, readOnlyModels, err := hydrator(ctx, key, value); err != nil {
		return err
	} else {
		models[key] = model
		for _, readOnlyModel := range readOnlyModels {
			if readOnlyModel == nil {
				continue
			}
			key, load := readOnlyModel()
			if load == nil {
				continue
			}
			if model, ok := models[key]; ok {
				load(model)
				continue
			}
			if err := loadModels(ctx, store, hydrator, key, models); err != nil {
				return err
			}
			load(models[key])
		}
		return nil
	}
}

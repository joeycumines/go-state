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
	"errors"
)

func OptionConfig(config Config) Option {
	return func(c *Config) error {
		*c = config
		return nil
	}
}

func OptionInit(init Init) Option {
	return func(config *Config) error {
		config.Init = init
		return nil
	}
}

func OptionUpdate(update Update) Option {
	return func(config *Config) error {
		config.Update = update
		return nil
	}
}

func OptionView(view View) Option {
	return func(config *Config) error {
		config.View = view
		return nil
	}
}

func OptionProducer(producer Producer) Option {
	return func(config *Config) error {
		config.Producer = producer
		return nil
	}
}

func OptionConsumer(consumer Consumer) Option {
	return func(config *Config) error {
		config.Consumer = consumer
		return nil
	}
}

func OptionReplay(replay func() bool) Option {
	return func(config *Config) error {
		config.Replay = replay
		return nil
	}
}

func OptionFetcher(fetcher Fetcher) Option {
	return func(config *Config) error {
		config.Fetcher = fetcher
		return nil
	}
}

func OptionStore(store Store) Option {
	return func(config *Config) error {
		config.Store = store
		return nil
	}
}

func OptionHydrator(hydrator Hydrator) Option {
	return func(config *Config) error {
		config.Hydrator = hydrator
		return nil
	}
}

func OptionDehydrator(dehydrator Dehydrator) Option {
	return func(config *Config) error {
		config.Dehydrator = dehydrator
		return nil
	}
}

func OptionKey(key string) Option {
	return func(config *Config) error {
		config.Key = key
		return nil
	}
}

func RunValidator(config *Config) error {
	if config.Init == nil {
		return errors.New("state.RunValidator nil init")
	}
	if config.Update == nil {
		return errors.New("state.RunValidator nil update")
	}
	if config.View == nil {
		return errors.New("state.RunValidator nil view")
	}
	if config.Producer == nil {
		return errors.New("state.RunValidator nil producer")
	}
	if config.Consumer == nil {
		return errors.New("state.RunValidator nil consumer")
	}
	return nil
}

func RunOptions(
	init Init,
	update Update,
	view View,
	producer Producer,
	consumer Consumer,
) []Option {
	return []Option{
		OptionInit(init),
		OptionUpdate(update),
		OptionView(view),
		OptionProducer(producer),
		OptionConsumer(consumer),
	}
}

func BatchValidator(config *Config) error {
	if err := RunValidator(config); err != nil {
		return err
	}
	if config.Fetcher == nil {
		return errors.New("state.BatchValidator nil fetcher")
	}
	return nil
}

func BatchOptions(
	init Init,
	update Update,
	view View,
	producer Producer,
	consumer Consumer,
	fetcher Fetcher,
) []Option {
	return append(
		RunOptions(
			init,
			update,
			view,
			producer,
			consumer,
		),
		OptionFetcher(fetcher),
	)
}

func AggregateValidator(config *Config) error {
	if err := BatchValidator(config); err != nil {
		return err
	}
	if config.Store == nil {
		return errors.New("state.AggregateValidator nil store")
	}
	if config.Hydrator == nil {
		return errors.New("state.AggregateValidator nil hydrator")
	}
	if config.Dehydrator == nil {
		return errors.New("state.AggregateValidator nil dehydrator")
	}
	return nil
}

func AggregateOptions(
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
) []Option {
	return append(
		BatchOptions(
			init,
			update,
			view,
			producer,
			consumer,
			fetcher,
		),
		OptionStore(store),
		OptionHydrator(hydrator),
		OptionDehydrator(dehydrator),
		OptionKey(key),
	)
}

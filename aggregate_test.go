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
	"github.com/joeycumines/go-bigbuff"
	"time"
	"github.com/go-test/deep"
	"encoding/json"
	"sync"
	"strings"
)

func TestAggregate_nilCtx(t *testing.T) {
	err := Aggregate(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		"",
	)
	if err == nil || err.Error() != "state.Aggregate nil ctx" {
		t.Fatal("unexpected err", err)
	}
}

func TestAggregate_nilUpdate(t *testing.T) {
	err := Aggregate(
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
		nil,
		nil,
		nil,
		"",
	)
	if err == nil || err.Error() != "state.Aggregate config error: nil update" {
		t.Fatal("unexpected err", err)
	}
}

func TestAggregate_nilFetcher(t *testing.T) {
	err := Aggregate(
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
		nil,
		nil,
		nil,
		"",
	)
	if err == nil || err.Error() != "state.Aggregate config error: nil fetcher" {
		t.Fatal("unexpected err", err)
	}
}

func TestAggregate_nilStore(t *testing.T) {
	err := Aggregate(
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
		func(ctx context.Context) ([]interface{}, bool, error) {
			panic("not implemented")
		},
		nil,
		nil,
		nil,
		"",
	)
	if err == nil || err.Error() != "state.Aggregate config error: nil store" {
		t.Fatal("unexpected err", err)
	}
}

func TestAggregate_counter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), 1, 2))
	defer cancel()

	var (
		ticker     exampleTicker
		store      = new(mockStore)
		storeToMap = func() map[string]exampleTicker {
			result := make(map[string]exampleTicker)

			(*sync.Map)(store).Range(func(key, value interface{}) bool {
				var m exampleTicker

				if err := json.Unmarshal(value.([]byte), &m); err != nil {
					t.Fatal(err)
				}

				result[key.(string)] = m

				return true
			})

			return result
		}
		hydrator Hydrator = func(ctx context.Context, key string, value []byte) (model interface{}, readOnlyModels []func() (key string, load func(model interface{})), err error) {
			if ctx == nil || 2 != ctx.Value(1) {
				t.Error("unexpected ctx", ctx)
			}

			if key != "some_key" {
				t.Error("unexpected key", key)
			}

			var m exampleTicker

			err = json.Unmarshal(value, &m)

			if err != nil {
				return
			}

			return m, []func() (key string, load func(model interface{})){
				func() (key string, load func(model interface{})) {
					return "", nil
				},
				nil,
				func() (key string, load func(model interface{})) {
					return "", nil
				},
			}, nil
		}
		dehydrator Dehydrator = func(ctx context.Context, key string, model interface{}) (value []byte, err error) {
			if ctx == nil || 2 != ctx.Value(1) {
				t.Error("unexpected ctx", ctx)
			}

			if key != "some_key" {
				t.Error("unexpected key", key)
			}

			return json.Marshal(model)
		}
		key = "some_key"
	)

	doAgg := func() {
		buffer := new(bigbuff.Buffer)

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

		err = Aggregate(
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
			store,
			hydrator,
			dehydrator,
			key,
		)

		if err != nil {
			t.Fatal(err)
		}

		if !exampleTargetTickerCalled {
			t.Error("expected target commands to be called")
		}
	}

	doAgg()

	if diff := deep.Equal(ticker, exampleTicker{
		HorizontalCounter: -7,
		VerticalCounter:   7,
		HorizontalMulti:   91239,
		VerticalMulti:     23111,
	}); diff != nil {
		t.Error("unexpected ticker", diff)
	}

	// check that the data store matches...
	m := storeToMap()

	if diff := deep.Equal(m, map[string]exampleTicker{
		key: {
			HorizontalCounter: -7,
			VerticalCounter:   7,
			HorizontalMulti:   91239,
			VerticalMulti:     23111,
		},
	}); diff != nil {
		t.Fatal("unexpected diff", diff)
	}

	// hydrate run it again, it will start at the previous model...
	doAgg()

	if diff := deep.Equal(ticker, exampleTicker{
		HorizontalCounter: -14,
		VerticalCounter:   14,
		HorizontalMulti:   91239,
		VerticalMulti:     23111,
	}); diff != nil {
		t.Error("unexpected ticker", diff)
	}

	// check that the data store matches...
	m = storeToMap()

	if diff := deep.Equal(m, map[string]exampleTicker{
		key: {
			HorizontalCounter: -14,
			VerticalCounter:   14,
			HorizontalMulti:   91239,
			VerticalMulti:     23111,
		},
	}); diff != nil {
		t.Fatal("unexpected diff", diff)
	}
}

type (
	mA struct {
		Bob int
	}
	mB struct {
		D   string
		Wat string
		Dad *mD
	}
	mC struct {
		A   string
		B   string
		Aww *mA
		Bad *mB
	}
	mD struct {
		C   string
		Cow *mC
	}
)

func TestLoadModels(t *testing.T) {
	// do not ask why I did this

	m := new(sync.Map)

	var (
		hydrator Hydrator = func(ctx context.Context, key string, value []byte) (model interface{}, readOnlyModels []func() (key string, load func(model interface{})), err error) {
			switch strings.Split(key, "/")[0] {
			case `A`:
				a := new(mA)
				err = json.Unmarshal(value, a)
				if err != nil {
					t.Fatal(err)
				}
				model = a
				return
			case `B`:
				b := new(mB)
				err = json.Unmarshal(value, b)
				if err != nil {
					t.Fatal(err)
				}
				model = b
				readOnlyModels = append(
					readOnlyModels,
					func() (key string, load func(model interface{})) {
						return `D/` + b.D, func(model interface{}) {
							if model == nil {
								return
							}
							b.Dad = model.(*mD)
						}
					},
				)
				return
			case `C`:
				c := new(mC)
				err = json.Unmarshal(value, c)
				if err != nil {
					t.Fatal(err)
				}
				model = c
				readOnlyModels = append(
					readOnlyModels,
					func() (key string, load func(model interface{})) {
						return `A/` + c.A, func(model interface{}) {
							if model == nil {
								return
							}
							c.Aww = model.(*mA)
						}
					},
					func() (key string, load func(model interface{})) {
						return `B/` + c.B, func(model interface{}) {
							if model == nil {
								return
							}
							c.Bad = model.(*mB)
						}
					},
				)
				return
			case `D`:
				d := new(mD)
				err = json.Unmarshal(value, d)
				if err != nil {
					t.Fatal(err)
				}
				model = d
				readOnlyModels = append(
					readOnlyModels,
					func() (key string, load func(model interface{})) {
						return `C/` + d.C, func(model interface{}) {
							if model == nil {
								return
							}
							d.Cow = model.(*mC)
						}
					},
				)
				return
			default:
				t.Fatal("unexpected key", key)
				return
			}
		}
	)

	model, err := loadModel(
		context.Background(),
		(*mockStore)(m),
		hydrator,
		"some_key",
	)

	if err != nil {
		t.Fatal(err)
	}

	if model != nil {
		t.Fatal("expected nil model")
	}
	/*
		mA struct {
			Bob int
		}
		mB struct {
			D   string
			Wat string
			Dad *mD
		}
		mC struct {
			A   string
			B   string
			Aww *mA
			Bad *mB
		}
		mD struct {
			C   string
			Cow *mC
		}
	 */
	m.Store(
		`C/1`,
		[]byte(`
			{
				"A": "2",
				"B": "3"
			}
		`),
	)
	model, err = loadModel(
		context.Background(),
		(*mockStore)(m),
		hydrator,
		"C/1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(model, &mC{
		A: "2",
		B: "3",
	}); diff != nil {
		t.Fatal("expected diff", diff)
	}
	m.Store(
		`A/2`,
		[]byte(`
			{
				"Bob": 2131
			}
		`),
	)
	model, err = loadModel(
		context.Background(),
		(*mockStore)(m),
		hydrator,
		"C/1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(model, &mC{
		A: "2",
		B: "3",
		Aww: &mA{
			Bob: 2131,
		},
	}); diff != nil {
		t.Fatal("expected diff", diff)
	}
	m.Store(
		`B/3`,
		[]byte(`
			{
				"D": "4",
				"Wat": "F",
				"Dad": {
					"C": "not real"
				}
			}
		`),
	)
	model, err = loadModel(
		context.Background(),
		(*mockStore)(m),
		hydrator,
		"C/1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(model, &mC{
		A: "2",
		B: "3",
		Aww: &mA{
			Bob: 2131,
		},
		Bad: &mB{
			D:   "4",
			Wat: "F",
			Dad: &mD{
				C: "not real",
			},
		},
	}); diff != nil {
		t.Fatal("expected diff", diff)
	}
	m.Store(
		`D/4`,
		[]byte(`
			{
				"C": "1"
			}
		`),
	)
	model, err = loadModel(
		context.Background(),
		(*mockStore)(m),
		hydrator,
		"C/1",
	)
	if err != nil {
		t.Fatal(err)
	}
	c := model.(*mC)
	cPtr := c.Bad.Dad.Cow
	if cPtr != c {
		t.Fatal("unexpected ptr", cPtr)
	}
	c.Bad.Dad.Cow = nil
	if diff := deep.Equal(c, &mC{
		A: "2",
		B: "3",
		Aww: &mA{
			Bob: 2131,
		},
		Bad: &mB{
			D:   "4",
			Wat: "F",
			Dad: &mD{
				C: "1",
			},
		},
	}); diff != nil {
		t.Fatal("expected diff", diff)
	}
}

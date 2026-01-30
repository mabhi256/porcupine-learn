package main

import (
	"fmt"
	"math/rand/v2"
	"sync"

	"github.com/anishathalye/porcupine"
)

type TestConfig struct {
	NumClients   int
	OpsPerClient int
	Keys         []string
	UsePartition bool
}

func runTestForClient(config TestConfig, clientId int, store *KVStore, recorder *EventRecorder) {
	for i := range config.OpsPerClient {
		key := config.Keys[rand.IntN(len(config.Keys))]

		if rand.IntN(2) == 0 {
			// SET operation
			value := fmt.Sprintf("v%d_%d", clientId, i)
			input := KVInput{Op: "set", Key: key, Value: value}

			id := recorder.RecordCall(clientId, input)
			store.Set(key, value)
			recorder.RecordReturn(clientId, id, KVOutput{Value: value})
		} else {
			// GET operation
			input := KVInput{Op: "get", Key: key}

			id := recorder.RecordCall(clientId, input)
			var value string
			value, _ = store.Get(key)
			recorder.RecordReturn(clientId, id, KVOutput{Value: value})
		}
	}
}

func runTest(config TestConfig) (porcupine.CheckResult, porcupine.LinearizationInfo) {
	store := NewKVStore()
	recorder := NewRecorder()

	var wg sync.WaitGroup

	for clientId := range config.NumClients {
		wg.Go(func() {
			runTestForClient(config, clientId, store, recorder)
		})
	}

	wg.Wait()

	events := recorder.GetEvents()

	var model porcupine.Model
	if config.UsePartition {
		model = storeModel
	} else {
		model = naiveModel
	}

	return porcupine.CheckEventsVerbose(model, events, 0)
}

package main

import (
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sync/atomic"

	"github.com/anishathalye/porcupine"
)

type KVInput struct {
	Op    string // "get" or "set"
	Key   string
	Value string
}

type KVOutput struct {
	Value string
}

var naiveModel = porcupine.Model{
	Init: func() any {
		return map[string]string{}
	},
	Step: func(state, input, output any) (bool, any) {
		curr := state.(map[string]string)
		inp := input.(KVInput)
		out := output.(KVOutput)

		// Idealized sequential behaviour
		switch inp.Op {
		case "set":
			// create new state
			newState := make(map[string]string)
			maps.Copy(newState, curr)
			// newState := maps.Clone(curr) // shallow clone, will not work for map[string]*SomeType
			newState[inp.Key] = inp.Value
			return true, newState

		case "get":
			return out.Value == curr[inp.Key], curr

		default:
			panic("Unexpected operation")
		}
	},
	DescribeOperation: func(input, output any) string {
		inp := input.(KVInput)
		out := output.(KVOutput)

		switch inp.Op {
		case "set":
			return fmt.Sprintf("put('%s', '%s')", inp.Key, inp.Value)
		case "get":
			return fmt.Sprintf("get('%s') -> '%s'", inp.Key, out.Value)
		default:
			panic("Unexpected operation")
		}
	},
	// state1 := map[string]string{"x": "1"}
	// state2 := map[string]string{"x": "1"}
	// state1 == state2  // COMPILE ERROR! Maps are not comparable
	Equal: func(state1, state2 any) bool {
		return reflect.DeepEqual(state1, state2)
	},
}

// Without partition, porcupine potentially tries all possible orderings for a concurrent operation
// If there are 5 concurrent operations in the kvstore
// - Porcupine may try 5! (120) orderings to check if a valid sequence of operations exist
// - This is dreadfully slow
//
// With partitioning we can check each key with an independent history
// - Instead of 5, we have say 2 ops for key x and 3 for y
// - Porcupine may check 2! + 3! (2+6=8) possible orderings
var storeModel = porcupine.Model{
	// history = [put(x, "1"), put(y, "2"), get(x) → "1", get(y) → "2", put(x, "3"), get(x) → "3"]
	// partition = [
	// 	 [put(x, "1"), get(x) → "1", put(x, "3"), get(x) → "3"],
	//   [put(y, "2"), get(y) → "2"]
	// ]
	Partition: func(history []porcupine.Operation) [][]porcupine.Operation {
		keyMap := make(map[string][]porcupine.Operation)
		for _, v := range history {
			inp := v.Input.(KVInput)
			keyMap[inp.Key] = append(keyMap[inp.Key], v)
		}

		// keys := make([]string, len(keyMap))
		// for key := range keyMap {
		// 	keys = append(keys, key)
		// }
		// slices.Sort(keys)

		partition := make([][]porcupine.Operation, 0, len(keyMap))
		for _, v := range keyMap {
			partition = append(partition, v)
		}
		return partition
	},
	PartitionEvent: func(history []porcupine.Event) [][]porcupine.Event {
		keyMap := make(map[string][]porcupine.Event)
		idMap := make(map[int]string) // Id -> Key

		for _, event := range history {
			if event.Kind == porcupine.CallEvent {
				inp := event.Value.(KVInput)
				keyMap[inp.Key] = append(keyMap[inp.Key], event)
				idMap[event.Id] = inp.Key
			} else {
				key := idMap[event.Id]
				keyMap[key] = append(keyMap[key], event)
			}
		}

		partition := make([][]porcupine.Event, 0, len(keyMap))
		for _, v := range keyMap {
			partition = append(partition, v)
		}
		return partition
	},
	Init: func() any {
		// Each partition is a single key. So init is empty string
		return ""
	},
	Step: func(state, input, output any) (bool, any) {
		curr := state.(string)
		inp := input.(KVInput)
		out := output.(KVOutput)

		switch inp.Op {
		case "set":
			return true, inp.Value
		case "get":
			return out.Value == curr, curr
		default:
			panic("Unexpected operation")
		}
	},
	DescribeOperation: func(input, output any) string {
		inp := input.(KVInput)
		out := output.(KVOutput)

		switch inp.Op {
		case "set":
			return fmt.Sprintf("put('%s', '%s')", inp.Key, inp.Value)
		case "get":
			return fmt.Sprintf("get('%s') -> '%s'", inp.Key, out.Value)
		default:
			panic("Unexpected operation")
		}
	},
}

type EventRecorder struct {
	events []porcupine.Event
	nextId atomic.Uint32
}

func NewRecorder() *EventRecorder {
	return &EventRecorder{}
}

func (er *EventRecorder) RecordCall(clientId int, input KVInput) int {
	nextId := er.nextId.Add(1)
	id := int(nextId - 1)

	er.events = append(er.events, porcupine.Event{
		ClientId: clientId,
		Kind:     porcupine.CallEvent,
		Value:    input,
		Id:       id,
	})
	return id
}

func (er *EventRecorder) RecordReturn(clientId int, id int, output KVOutput) {
	er.events = append(er.events, porcupine.Event{
		ClientId: clientId,
		Kind:     porcupine.ReturnEvent,
		Value:    output,
		Id:       id,
	})
}

func (er *EventRecorder) GetEvents() []porcupine.Event {
	return slices.Clone(er.events)
}

package main

import (
	"fmt"
	"os"

	"github.com/anishathalye/porcupine"
)

type RegisterInput struct {
	Op    string // "get" or "set"
	Value int    // value to set (ignored for get)
}

func main() {
	// Define the idealized sequential specification
	registerModel := porcupine.Model{
		// Initial state: register holds 0
		Init: func() any {
			return 0
		},
		// Step: given state + input + output, is this valid?
		Step: func(state, input, output any) (bool, any) {
			reg := state.(int)
			inp := input.(RegisterInput)
			out := output.(int)

			switch inp.Op {
			case "set":
				return true, inp.Value // Always succeeds, new state is the value
			case "get":
				return out == reg, reg // Check if output matches current state
			default:
				panic("Unexpected operation")
			}
		},
		DescribeOperation: func(input, output any) string {
			inp := input.(RegisterInput)
			switch inp.Op {
			case "set":
				return fmt.Sprintf("put('%d')", inp.Value)
			case "get":
				return fmt.Sprintf("get() -> '%d'", output.(int))
			default:
				panic("Unexpected operation")
			}
		},
	}

	// [Client 0] t0-t1: set(10)
	// [Client 0] t2-t3: get() → 10
	// [Client 1] t4-t5: set(20)
	// [Client 0] t6-t7: get() → 20

	// Call = start time, Return = end time
	ops1 := []porcupine.Operation{
		{ClientId: 0, Input: RegisterInput{"set", 10}, Call: 0, Output: 10, Return: 1},
		{ClientId: 0, Input: RegisterInput{"get", 0}, Call: 2, Output: 10, Return: 3},
		{ClientId: 1, Input: RegisterInput{"set", 20}, Call: 4, Output: 20, Return: 5},
		{ClientId: 0, Input: RegisterInput{"get", 0}, Call: 6, Output: 20, Return: 7},
	}

	result, info := porcupine.CheckOperationsVerbose(registerModel, ops1, 0)

	if result == porcupine.Ok {
		fmt.Println("1. ✅ LINEARIZABLE")
	} else {
		fmt.Println("1. ❌ NOT LINEARIZABLE")
	}

	saveVisualization(registerModel, info, "01_basic.html")
	fmt.Println()

	//------------------------

	// [Client 0] t0-t1: set(100)
	// [Client 1] t2-t3: set(200)
	// [Client 0] t4-t5: get() → 100 (STALE READ)
	// The get() happens AFTER set(200), so it MUST return 200.

	ops2 := []porcupine.Operation{
		{ClientId: 0, Input: RegisterInput{"set", 100}, Call: 0, Output: 100, Return: 1},
		{ClientId: 1, Input: RegisterInput{"set", 200}, Call: 2, Output: 200, Return: 3},
		{ClientId: 0, Input: RegisterInput{"get", 0}, Call: 4, Output: 100, Return: 5},
	}

	result, info = porcupine.CheckOperationsVerbose(registerModel, ops2, 0)

	if result == porcupine.Ok {
		fmt.Println("2. ✅ LINEARIZABLE")
	} else {
		fmt.Println("2. ❌ NOT LINEARIZABLE")
	}

	saveVisualization(registerModel, info, "02_invalid.html")
	fmt.Println()

	// -------------------

	// t=0-----------5---6---7
	// |---set(10)---|
	//     |---set(20)---|
	//         |----get()----|
	//

	testValues := []int{0, 10, 20, 30}
	explaination := []string{
		"get()→0:  get() happens first, before any set",
		"get()→10: set(10) → get() → set(20)",
		"get()→20: set(20) → get() → set(10)",
		"get()→30: Impossible! No operation sets 30",
	}

	for i, val := range testValues {
		ops := []porcupine.Operation{
			{ClientId: 0, Input: RegisterInput{"set", 10}, Call: 0, Output: 10, Return: 5},
			{ClientId: 1, Input: RegisterInput{"set", 20}, Call: 1, Output: 20, Return: 6},
			{ClientId: 2, Input: RegisterInput{"get", 0}, Call: 2, Output: val, Return: 7},
		}

		result, info := porcupine.CheckOperationsVerbose(registerModel, ops, 0)

		fmt.Printf("3(%d). %s\n", i, explaination[i])
		if result == porcupine.Ok {
			fmt.Println("✅ VALID")
		} else {
			fmt.Println("❌ INVALID")
		}
		saveVisualization(registerModel, info, fmt.Sprintf("03_concurrent_get_%d.html", val))
		fmt.Println()
	}

	// ------------

	events := []porcupine.Event{
		// Invocation / Completion Events in chronological order
		// Id links a Call/Return pair for the same ClientId — must match exactly once each
		{ClientId: 0, Kind: porcupine.CallEvent, Value: RegisterInput{"set", 100}, Id: 12341234},
		{ClientId: 1, Kind: porcupine.CallEvent, Value: RegisterInput{"get", 0}, Id: 56785678},
		{ClientId: 1, Kind: porcupine.ReturnEvent, Value: 100, Id: 56785678},
		{ClientId: 0, Kind: porcupine.ReturnEvent, Value: 100, Id: 12341234},
	}

	// t=--------------------
	// |-----set(100)-----|
	//    |---get()---|

	result, info = porcupine.CheckEventsVerbose(registerModel, events, 0)
	if result == porcupine.Ok {
		fmt.Println("4. ✅ VALID")
	} else {
		fmt.Println("4. ❌ INVALID")
	}
	saveVisualization(registerModel, info, "04_events.html")
}

func saveVisualization(model porcupine.Model, info porcupine.LinearizationInfo, filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	porcupine.Visualize(model, info, f)
	fmt.Printf("- Visualization saved to %s\n", filename)
}

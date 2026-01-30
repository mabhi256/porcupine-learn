package main

import (
	"fmt"
	"os"

	"github.com/anishathalye/porcupine"
)

// Linearizability Checking is in NP (Given a history, we just run it in order and verify in O(n) time)
// Linearizability Checking is NP-Hard. A known NP-Hard problem (Subset Sum) reduces to Linearizability Checking.
// Subset Sum - Is there a subset of S = {s₁, s₂, ..., sₙ} that sums to exactly t

type AdderInput struct {
	Op    string // "add" or "get"
	Value int
}

func main() {
	adderModel := porcupine.Model{
		Init: func() any { return 0 },
		Step: func(state, input, output any) (bool, any) {
			total := state.(int)
			inp := input.(AdderInput)
			out := output.(int)

			switch inp.Op {
			case "add":
				return true, total + inp.Value
			case "get":
				return out == total, total
			default:
				panic("Unexpected operation")
			}
		},
	}

	fmt.Println("Timeline:")
	fmt.Println("  t=0-----------------t=10")
	fmt.Println("  |------Add(3)-------|")
	fmt.Println("  |------Add(7)-------|")
	fmt.Println("  |------Add(5)-------|")
	fmt.Println("  |------Add(2)-------|")
	fmt.Println("  |------Get()→8------|")
	fmt.Println()

	addOps := []porcupine.Operation{
		{ClientId: 0, Input: AdderInput{"add", 3}, Call: 0, Output: 0, Return: 10},
		{ClientId: 1, Input: AdderInput{"add", 7}, Call: 0, Output: 0, Return: 10},
		{ClientId: 2, Input: AdderInput{"add", 5}, Call: 0, Output: 0, Return: 10},
		{ClientId: 3, Input: AdderInput{"add", 2}, Call: 0, Output: 0, Return: 10},
	}

	testValues := []int{0, 5, 7, 20}

	for _, v := range testValues {
		getOp := porcupine.Operation{ClientId: 4, Input: AdderInput{"get", 0}, Call: 0, Output: v, Return: 10}
		ops := append(addOps, getOp)

		result, info := porcupine.CheckOperationsVerbose(adderModel, ops, 0)
		if result == porcupine.Ok {
			fmt.Printf("Subset sum get()->%d is ✅ VALID\n", v)
		} else {
			fmt.Printf("Subset sum get()->%d is ❌ INVALID\n", v)
		}
		saveVisualization(adderModel, info, fmt.Sprintf("subset_sum_get_%d.html", v))
		fmt.Println()
	}
}

func saveVisualization(model porcupine.Model, info porcupine.LinearizationInfo, filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	porcupine.Visualize(model, info, f)
	fmt.Printf("- Visualization saved to %s\n", filename)
}

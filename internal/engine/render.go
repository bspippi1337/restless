package engine

import "fmt"

func RenderDOT(dot string) {
	fmt.Println(dot)
}

func Print(r *Result) {
	PrintResult(r)
}

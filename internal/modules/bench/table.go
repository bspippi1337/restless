package bench

import "fmt"

func PrintTable(r Result) {
	fmt.Println("==== BENCH RESULT ====")
	fmt.Printf("Total     : %d\n", r.TotalRequests)
	fmt.Printf("Errors    : %d\n", r.Errors)
	fmt.Printf("Duration  : %d ms\n", r.DurationMs)
	fmt.Printf("P50       : %d ms\n", r.P50Ms)
	fmt.Printf("P95       : %d ms\n", r.P95Ms)
	fmt.Printf("P99       : %d ms\n", r.P99Ms)
}

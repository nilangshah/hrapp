package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

// from fib_test.go
func BenchmarkFetch(bb *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < bb.N; n++ {
		client := createHrAppClient(svcAddr)
		wait.Add(1)
		fetch(1, client)
		wait.Wait()
	}
	ans := printResult(1)
	_, err := json.Marshal(ans)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(string(b))
}

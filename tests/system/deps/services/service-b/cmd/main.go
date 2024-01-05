package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Printf("service-b started.\n")
	resp, err := http.Get("http://localhost:9123/ops/liveness")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Response: %#v\n", resp)
	if resp.StatusCode != 200 {
		panic("wrong status code")
	}
	fmt.Printf("service-b done.\n")
}

package main

import (
	"sync"

	f "./factory"
)

func main() {
	var wg sync.WaitGroup

	rds := f.AWSFactory("rds")
	elasti := f.AWSFactory("elasti")

	wg.Add(2)

	go rds.List(&wg)

	go elasti.List(&wg)

	wg.Wait()
}

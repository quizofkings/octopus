package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:9090",
		Password: "", // no password set
		DB:       0,  // use default DB
		// ReadTimeout: 5 * time.Second,
		// PoolSize: 5,
	})

	rw := sync.Mutex{}
	count := 0
	wg := &sync.WaitGroup{}
	for i := 1; i <= 60; i++ {
		if i%30 == 0 {
			time.Sleep(2 * time.Second)
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			rw.Lock()
			count++
			rw.Unlock()
			fmt.Println(client.Ping().Result())
		}(wg, i)
	}

	wg.Wait()
}

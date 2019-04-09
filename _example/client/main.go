package main

import (
	"fmt"
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

RETRY:
	count := 0
	for i := 1; i <= 30; i++ {
		if i%30 == 0 {
			time.Sleep(2 * time.Second)
		}
		go func(i int) {
			fmt.Println(client.Ping().Result())
		}(i)
	}

	fmt.Println(count)
	goto RETRY
}

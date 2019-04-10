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
	for i := 1; i <= 100; i++ {
		if i%40 == 0 {
			time.Sleep(1 * time.Second)
		}
		go func(i int) {
			fmt.Println(client.Ping().Result())
			// fmt.Println(client.SMembers("peste:category:6").Result())
		}(i)
	}

	fmt.Println(count)
	goto RETRY
}

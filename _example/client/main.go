package main

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:9090",
		Password: "", // no password set
		DB:       0,  // use default DB
		// PoolSize: 20,
	})

	// load
	wg := &sync.WaitGroup{}
	for i := 0; i <= 50; i++ {
		// if i%30 == 0 {
		// 	time.Sleep(3 * time.Second)
		// }
		wg.Add(1)
		go func(i int, wg *sync.WaitGroup) {
			// pong, err := client.Ping().Result()
			// fmt.Println(i, pong, err)
			// ...
			fmt.Println(client.Get("challenge:xxx").Result())
			wg.Done()
		}(i, wg)
	}

	wg.Wait()
	fmt.Println("Done")

	// others...
	// fmt.Println("RUN")
	// fmt.Println(client.Get("challenge:xxx").Result())
	// fmt.Println(client.SMembers("pigfog:ssss").Result())
}

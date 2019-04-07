package main

import (
	"fmt"

	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:9090",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	// others...
	// fmt.Println(client.Get("pigfog:amir").Result())
	// fmt.Println(client.SMembers("pigfog:ssss").Result())
}

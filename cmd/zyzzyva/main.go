package main

import (
	"github.com/myl7/zyzzyva/pkg/client"
	"github.com/myl7/zyzzyva/pkg/conf"
	"github.com/myl7/zyzzyva/pkg/server"
	"sync"
)

func main() {
	err := conf.InitKeys()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for i := 0; i < conf.N; i++ {
		s := server.NewServer(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Run()
			if err != nil {
				panic(err)
			}
		}()
	}

	for i := 0; i < conf.M; i++ {
		c := client.NewClient(i + conf.N)
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.Run()
			if err != nil {
				panic(err)
			}
		}()
	}

	wg.Wait()
}

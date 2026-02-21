package syncx

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func WaitUntilSignaled() {
	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	select {
	case <-sigChan:
		break
	}
}

func WaitWork(worker func(), num int) {
	wg := sync.WaitGroup{}
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker()
		}()
	}
	wg.Wait()
}

package chanx

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestAdaptiveChan(t *testing.T) {
	ch := NewAdaptiveChan(100, 100, 100)

	var count int

	for i := 1; i < 200; i++ {
		ch.In() <- i
	}

	for i := 1; i < 60; i++ {
		v, _ := <-ch.Out()
		count += v.(int)
	}

	assert.Equal(t, 100, ch.buffer.Cap())

	for i := 200; i <= 1200; i++ {
		ch.In() <- i
	}
	assert.Equal(t, 1600, ch.buffer.Cap())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch.Out() {
			count += v.(int)
		}
	}()

	close(ch.In())

	wg.Wait()

	assert.Equal(t, 720600, count)

}

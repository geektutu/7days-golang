package singleflight

import (
	"sync"
	"testing"
	"time"
)

func TestDo(t *testing.T) {
	var g Group

	v, err := g.Do("key", func() (interface{}, error) {
		return "bar", nil
	})

	if v != "bar" || err != nil {
		t.Errorf("Do v = %v, error = %v", v, err)
	}
}

func TestGoDo(t *testing.T) {
	var g Group
	wg := sync.WaitGroup{}

	times := 10
	wg.Add(times)
	for i := 0; i < times; i++ {
		go func() interface{} {
			v, err := g.Do("key", func() (interface{}, error) {
				time.Sleep(time.Millisecond * 50)
				return "bar", nil
			})

			wg.Done()

			if v != "bar" || err != nil {
				t.Errorf("Do v = %v, error = %v", v, err)
			}

			return 1
		}()
	}

	wg.Wait()
}

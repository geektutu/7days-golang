package singleflight

import (
	"sync"
	"testing"
	"time"
)

// Record the fn execution times
var ExecNum = 0

func TestDoCase1(t *testing.T) {
	var ExecNum = 0
	var g Group
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			v, err := g.Do("key", func() (interface{}, error) {
				//If call fn once, add it once
				ExecNum++
				return "bar", nil
			})
			if v != "bar" || err != nil {
				t.Error("Exec Do err")
			}
		}()
	}

	//fn can be executed only one times
	if ExecNum != 1 {
		t.Error("singleFlight err")
	}
}

func TestDoCase2(t *testing.T) {
	var ExecNum = 0
	var g Group
	wg := sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 1000; i++ {
		go func() {
			v, err := g.Do("key", func() (interface{}, error) {
				//If call fn once, add it once
				ExecNum++
				return "bar", nil
			})
			if v != "bar" || err != nil {
				t.Error("Exec Do err")
			}
		}()
	}

	time.Sleep(time.Second)
	//One second later, fn will add one more
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			v, err := g.Do("key", func() (interface{}, error) {
				//If call fn once, add it once
				ExecNum++
				return "bar", nil
			})
			if v != "bar" || err != nil {
				t.Error("Exec Do err")
			}
		}()
	}

	if ExecNum != 2 {
		t.Error("singleFlight err")
	}
}

func TestDoCase3(t *testing.T) {
	var ExecNum = 0
	var g Group
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			v, err := g.Do("key", func() (interface{}, error) {
				//If call fn once, add it once
				ExecNum++
				return "bar", nil
			})
			if v != "bar" || err != nil {
				t.Error("Exec Do err")
			}

			//key1 != key，fn will add one more
			v, err = g.Do("key1", func() (interface{}, error) {
				//If call fn once, add it once
				ExecNum++
				return "bar1", nil
			})
			if v != "bar1" || err != nil {
				t.Error("Exec Do err")
			}
		}()
	}
	//fn can be executed only one times
	if ExecNum != 2 {
		t.Error("singleFlight err")
	}
}

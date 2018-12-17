package main

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestFanout(t *testing.T) {
	type in struct {
		a int
		b int
	}
	type out struct {
		name    string
		retsult int
	}

	evaluators := []Evaluator{
		EvaluatorFunc(func(i interface{}) (interface{}, error) {
			v := i.(in)
			return out{"adder", v.a + v.b}, nil
		}),
		EvaluatorFunc(func(i interface{}) (interface{}, error) {
			v := i.(in)
			return out{"subber", v.a - v.b}, nil
		}),
		EvaluatorFunc(func(i interface{}) (interface{}, error) {
			v := i.(in)
			time.Sleep(time.Duration(3) * time.Second)
			return out{"multiplier", v.a * v.b}, nil
		}),
		EvaluatorFunc(func(i interface{}) (interface{}, error) {
			v := i.(in)
			return out{"divider", v.a / v.b}, nil
		}),
	}

	rets, errs := DevideAndConquer(in{10, 1}, evaluators, time.Duration(2)*time.Second)
	log.Println(rets, errs)
}

func TestPool(t *testing.T) {
	p := NewPool(func() interface{} {
		return 0
	}, 3)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 2; j++ {
				p.Run(func(v interface{}) {
					log.Println("gopher", v.(int), "borrow item", i+v.(int))
					time.Sleep(time.Duration((i+1)*100) * time.Millisecond)
				})

			}
		}(i)
	}
	wg.Wait()
}

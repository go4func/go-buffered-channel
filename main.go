package main

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

type Evaluator interface {
	Evaluate(interface{}) (interface{}, error)
	Name() string
}
type EvaluatorFunc func(interface{}) (interface{}, error)

func (ef EvaluatorFunc) Evaluate(in interface{}) (interface{}, error) {
	return ef(in)
}
func (ef EvaluatorFunc) Name() string {
	return runtime.FuncForPC(reflect.ValueOf(ef).Pointer()).Name()
}

func DevideAndConquer(data interface{}, evaluators []Evaluator, timeout time.Duration) ([]interface{}, []error) {
	start := time.Now()
	gather := make(chan interface{}, len(evaluators))
	errors := make(chan error, len(evaluators))
	for _, v := range evaluators {
		go func(e Evaluator) {
			ch := make(chan interface{}, 1)
			ech := make(chan error, 1)
			go func() {
				ret, err := e.Evaluate(data)
				if err != nil {
					ech <- err
				} else {
					ch <- ret
				}
			}()
			select {
			case r := <-ch:
				gather <- r
			case e := <-ech:
				errors <- e
			case <-time.After(timeout):
				errors <- fmt.Errorf("%s timeout after %v with data %v", e.Name(), timeout, data)
			}

		}(v)
	}

	outs := make([]interface{}, 0, len(evaluators))
	errs := make([]error, 0, len(evaluators))
	for range evaluators {
		select {
		case r := <-gather:
			outs = append(outs, r)
		case e := <-errors:
			errs = append(errs, e)
		}
	}

	fmt.Println("enlapse time:", time.Now().Sub(start))

	return outs, errs
}

var f1 EvaluatorFunc = func(i interface{}) (interface{}, error) {
	return i.(int) * 11, nil
}

var f2 EvaluatorFunc = func(i interface{}) (interface{}, error) {
	time.Sleep(time.Duration(3) * time.Second)
	return 0, errors.New("error from f2")
}

type Factory func() interface{}
type Processor func(interface{})
type poolInner struct {
	items chan interface{}
}

type Pool interface {
	Run(Processor)
	RunWithTimeout(Processor, time.Duration) error
}

func NewPool(f Factory, count int) Pool {
	pI := &poolInner{items: make(chan interface{}, count)}
	for i := 0; i < count; i++ {
		pI.items <- f()
	}
	return pI
}

func (pi *poolInner) Run(p Processor) {
	item := <-pi.items
	defer func() {
		pi.items <- item
	}()
	p(item)
}

func (pi *poolInner) RunWithTimeout(p Processor, t time.Duration) error {
	select {
	case item := <-pi.items:
		defer func() {
			pi.items <- item
		}()
		p(item)
	case <-time.After(t):
		return fmt.Errorf("Process timeout after %v", t)
	}
	return nil
}

func main() {
}

package pool

import (
	"errors"
	"log"
	"sync/atomic"
)

type (
	Pool struct {
		capacity int32
		state    int32
		running  int32
		tasks    chan *Task
	}
	Task struct {
		Handler func(v ...interface{})
		Params  []interface{}
	}
)

const (
	Running = iota
	Stopped
)

func NewPool(capacity int32) (*Pool, error) {
	if capacity <= 0 {
		return nil, errors.New("capacity less than 0")
	}
	instance := &Pool{
		capacity: capacity,
		state:    Running,
		tasks:    make(chan *Task, capacity),
	}
	return instance, nil
}

func (p *Pool) Run() {
	atomic.AddInt32(&p.running, 1)
	go func() {
		defer func() {
			atomic.AddInt32(&p.running, -1)
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()
		for {
			select {
			case task, ok := <-p.tasks:
				if ok {
					task.Handler(task.Params)
				}
			}
		}
	}()
}

func (p *Pool) AddTask(task *Task) error {
	if p.state == Stopped {
		return errors.New("task pool is closed")
	}
	running := atomic.LoadInt32(&p.running)
	capacity := atomic.LoadInt32(&p.capacity)
	if running < capacity {
		p.Run()
	}
	p.tasks <- task
	return nil
}

func (p *Pool) Close() {
	p.state = Stopped
	for {
		if len(p.tasks) <= 0 {
			close(p.tasks)
			return
		}
	}
}

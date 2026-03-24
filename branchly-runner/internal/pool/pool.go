package pool

import "sync"

type Pool struct {
	sem chan struct{}
	wg  sync.WaitGroup
}

func New(maxConcurrent int) *Pool {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	return &Pool{sem: make(chan struct{}, maxConcurrent)}
}

func (p *Pool) TryGo(fn func()) bool {
	select {
	case p.sem <- struct{}{}:
		p.wg.Add(1)
		go func() {
			defer func() {
				<-p.sem
				p.wg.Done()
			}()
			fn()
		}()
		return true
	default:
		return false
	}
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

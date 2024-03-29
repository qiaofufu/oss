package singleflight

import "sync"

type operation struct {
	wg  sync.WaitGroup
	val any
	err error
}

type Flight struct {
	mu  sync.Mutex
	ops map[string]*operation
}

func (f *Flight) Do(key string, fn func() (any, error)) (any, error) {
	f.mu.Lock()
	if op, ok := f.ops[key]; ok {
		f.mu.Unlock()
		op.wg.Wait()
		return op.val, op.err
	}
	op := &operation{}
	op.wg.Add(1)
	f.ops[key] = op
	f.mu.Unlock()

	op.val, op.err = fn()
	op.wg.Done()

	f.mu.Lock()
	delete(f.ops, key)
	f.mu.Unlock()

	return op.val, op.err
}

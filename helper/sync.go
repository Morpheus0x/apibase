package helper

import "sync"

type MtxString struct {
	mtx  sync.Mutex
	data string
	x    string
}

func (s *MtxString) Set(data string) {
	s.mtx.Lock()
	s.data = data
	s.mtx.Unlock()
}

func (s *MtxString) Get() (data string) {
	s.mtx.Lock()
	data = s.data
	s.mtx.Unlock()
	return
}

package main

import (
	"sync"
	"sync/atomic"
)

type summary struct {
	sync.Mutex

	num uint32

	Received uint64
	Donated  uint64
	Faulty   uint64

	Top       [3]string
	TopAmount map[string]int64
}

func NewSummary() *summary {
	sum := summary{}
	sum.TopAmount = make(map[string]int64, 3)
	return &sum
}

func (s *summary) IncrementNum(amount int) {
	atomic.AddUint32(&s.num, uint32(amount))
}

func (s *summary) IncrementReceived(amount int64) {
	atomic.AddUint64(&s.Received, uint64(amount))
}

func (s *summary) IncrementDonated(amount int64) {
	atomic.AddUint64(&s.Donated, uint64(amount))
}

func (s *summary) IncrementFaulty(amount int64) {
	atomic.AddUint64(&s.Faulty, uint64(amount))
}

func (s *summary) UpdateTop(name string, amount int64) {
	s.Lock()
	defer s.Unlock()

	var old string
	for i := 2; i >= 0; i-- {
		if amount < s.TopAmount[s.Top[i]] {
			break
		}

		old = s.Top[i]
		s.Top[i] = name

		if i == 2 {
			s.TopAmount[name] = amount
			delete(s.TopAmount, old)
			continue
		}

		s.Top[i+1] = old
	}
}

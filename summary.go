package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type summary struct {
	sync.Mutex

	Num uint32

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
	atomic.AddUint32(&s.Num, uint32(amount))
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

func (s *summary) Print() {
	p := message.NewPrinter(language.English)

	fmt.Println()
	fmt.Printf("%20s: THB %20s\n", "total received", p.Sprintf("%d", s.Received))
	fmt.Printf("%20s: THB %20s\n", "successfully donated", p.Sprintf("%d", s.Donated))
	fmt.Printf("%20s: THB %20s\n", "faulty donation", p.Sprintf("%d", s.Faulty))
	fmt.Println()
	fmt.Printf("%20s: THB %20s\n", "average per person", p.Sprintf("%.2f", float64(s.Received)/float64(s.Num)))
	fmt.Printf("%20s: THB %s\n", "top donors", s.Top[0])
	fmt.Printf("%20s  THB %s\n", "", s.Top[1])
	fmt.Printf("%20s  THB %s\n", "", s.Top[2])
}

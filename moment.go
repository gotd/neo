package neo

import "time"

type moment struct {
	when time.Time
	do   func(time time.Time)
}

type moments []moment

func (m moments) do(time time.Time) {
	for _, doer := range m {
		doer.do(time)
	}
}

func (m moments) Len() int {
	return len(m)
}

func (m moments) Less(i, j int) bool {
	return m[i].when.Before(m[j].when)
}

func (m moments) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

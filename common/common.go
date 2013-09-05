package common

type Sequence int

func (s *Sequence) Next() Sequence {
	res := *s
	(*s)++
	return res
}

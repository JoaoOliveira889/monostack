package tui

type sortState struct {
	field     int
	ascending bool
}

func (s *sortState) Cycle(maxField int) {
	s.field = (s.field + 1) % (maxField + 1)
	s.ascending = !s.ascending
	if s.field == 0 {
		s.ascending = true
	}
}

func (s sortState) Field() int      { return s.field }
func (s sortState) Ascending() bool { return s.ascending }

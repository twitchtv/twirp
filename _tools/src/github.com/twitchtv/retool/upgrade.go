package main

func (s spec) upgrade(t *tool) {
	idx := s.find(t)
	if idx == -1 {
		log(t.Repository + " is not yet installed (did you mean retool add?)")
		return
	}

	s.Tools[idx] = t

	s.sync()

	err := s.write()
	if err != nil {
		fatal("unable to remove "+t.Repository, err)
	}
}

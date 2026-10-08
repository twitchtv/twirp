package main

func (s spec) add(t *tool) {
	if s.find(t) != -1 {
		log(t.Repository + " already installed (did you mean retool upgrade?)")
		return
	}

	s.Tools = append(s.Tools, t)

	s.sync()

	err := s.write()
	if err != nil {
		fatal("unable to add "+t.Repository, err)
	}
}

package main

func (s spec) remove(t *tool) {
	idx := s.find(t)
	if idx == -1 {
		fatal(t.Repository+" is not in tools.json", nil)
	}
	s.Tools = append(s.Tools[:idx], s.Tools[idx+1:]...)
	err := s.write()
	if err != nil {
		fatal("unable to remove "+t.Repository, err)
	}

	s.sync()
}

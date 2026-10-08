package main

// builds all tools in the spec file using whatever is installed in the tool directory (_tools,
// typically). Shouldn't do any network if _tools is set up correctly.
func (s spec) build() {
	err := setGoEnv()
	if err != nil {
		fatal("unable to set GOPATH and GOBIN env variables", err)
	}

	m := getManifest()
	for _, t := range s.Tools {
		err := install(t)
		if err != nil {
			fatalExec("go install "+t.Repository, err)
		}
	}
	m.replace(s.Tools)
	m.write()
}

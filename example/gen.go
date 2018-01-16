package example

//go:generate protoc -I . service.proto --twirp_out=. --go_out=. --python_out=. --twirp_python_out=.

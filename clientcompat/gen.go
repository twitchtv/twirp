package main

//go:generate protoc --proto_path=. --twirp_out=./internal/clientcompat --go_out=./internal/clientcompat clientcompat.proto
//go:generate protoc --proto_path=. --twirp_python_out=./pycompat --python_out=./pycompat clientcompat.proto

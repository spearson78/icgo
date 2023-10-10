.PHONY : example

example:
	go build -gcflags "-N -l" -o bin/example ./example/.
	./bin/example

all: example
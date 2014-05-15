
ifndef prefix
prefix=/usr/local
endif

PROGRAM_NAME=gauge

build:
	export GOPATH=$(GOPATH):`pwd` && cd src && go build && mv src ../$(PROGRAM_NAME)

build-debug:
	export GOPATH=$(GOPATH):`pwd` && cd src && go build -gcflags "-N -l" && mv src ../$(PROGRAM_NAME)

test:
	export GOPATH=$(GOPATH):`pwd` && cd src && go test

install:
	install -m 755 -d $(prefix)/bin
	install -m 755 $(PROGRAM_NAME) $(prefix)/bin
	install -m 755 -d $(prefix)/share/$(PROGRAM_NAME)/skel
	install -m 755 -d $(prefix)/share/$(PROGRAM_NAME)/skel/env
	install -m 644 skel/hello_world.spec $(prefix)/share/$(PROGRAM_NAME)/skel
	install -m 644 skel/default.properties $(prefix)/share/$(PROGRAM_NAME)/skel/env

	

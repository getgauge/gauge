
ifndef prefix
prefix=/usr/local
endif

PROGRAM_NAME=twist2

build:
	cd src && go build && mv src ../twist2

install:
	install -m 755 -d $(prefix)/bin
	install -m 755 twist2 $(prefix)/bin
	install -m 755 -d $(prefix)/share/$(PROGRAM_NAME)/skel
	install -m 644 skel/* $(prefix)/share/$(PROGRAM_NAME)/skel
	

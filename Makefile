
ifndef prefix
prefix=/usr/local
endif

build:
	cd src && go build && mv src ../twist2

install:
	install -m 755 -d $(prefix)/bin
	install -m 755 twist2 $(prefix)/bin
	

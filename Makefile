
ifndef prefix
prefix=/usr/local
endif

build:
	go run make.go

test:
	go run make.go --test

install:
	go run make.go --install --prefix $(prefix)


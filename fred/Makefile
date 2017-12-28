name=$(shell basename $(realpath .))
script=profit

all:
	go build
	install $(name) ${HOME}/bin
	install $(script) ${HOME}/bin

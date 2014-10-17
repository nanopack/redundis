deps:
	git submodule update --init
	make -C deps/luvit-redis

all: deps
	
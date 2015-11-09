.PHONY: all deps install install-deps install-libs

deps:
	git submodule update --init
	cp Makefile.luvit-redis deps/luvit-redis/Makefile
	make -C deps/luvit-redis

all: deps

install: install-deps install-libs
	mkdir -p ${DESTDIR}${PREFIX}/redundis
	cp redundis.lua ${DESTDIR}${PREFIX}/redundis/redundis.lua

install-deps:
	mkdir -p ${DESTDIR}${PREFIX}/redundis/deps/luvit-redis/build
	mkdir -p ${DESTDIR}${PREFIX}/redundis/deps/luvit-redis/lib
	for i in deps/luvit-redis/build/redis.luvit deps/luvit-redis/lib/commands.lua deps/luvit-redis/lib/init.lua deps/luvit-redis/package.lua; \
		do; \
		cp $$i ${DESTDIR}${PREFIX}/redundis/$$i; \
	done

install-libs:
	mkdir -p ${DESTDIR}${PREFIX}/redundis/lib
	cp lib/main.lua ${DESTDIR}${PREFIX}/redundis/lib/main.lua

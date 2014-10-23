.PHONY: all deps install install-deps install-libs

deps:
	git submodule update --init
	cp Makefile.luvit-redis deps/luvit-redis/Makefile
	make -C deps/luvit-redis

all: deps

install: install-deps install-libs
	mkdir -p ${DESTDIR}${PREFIX}/redis_proxy
	cp redis_proxy.lua ${DESTDIR}${PREFIX}/redis_proxy/redis_proxy.lua

install-deps:
	mkdir -p ${DESTDIR}${PREFIX}/redis_proxy/deps/luvit-redis/build
	mkdir -p ${DESTDIR}${PREFIX}/redis_proxy/deps/luvit-redis/lib
	for i in deps/luvit-redis/build/redis.luvit deps/luvit-redis/lib/commands.lua deps/luvit-redis/lib/init.lua deps/luvit-redis/package.lua; \
		cp $$i ${DESTDIR}${PREFIX}/$$i
	done

install-libs:
	mkdir -p ${DESTDIR}${PREFIX}/redis_proxy/libs
	cp libs/main.lua ${DESTDIR}${PREFIX}/redis_proxy/libs/main.lua
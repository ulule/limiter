.PHONY: test

cleandb:
	@(redis-cli KEYS "limitertests*" | xargs redis-cli DEL)

test: cleandb
	@(go test -v -run ^Test)

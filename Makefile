.PHONY: test

cleandb:
	@(redis-cli KEYS "limitertests:*" | xargs redis-cli DEL)

test: cleandb
	@(scripts/test)

lint:
	@(scripts/lint)

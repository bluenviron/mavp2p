
YMLS := $(shell ls *yml | sed 's/\.yml$$//')

help:
	@echo "usage: make [test]"
	@echo "tests: $(YMLS)"

.PHONY: $(YMLS)

$(YMLS):
	docker-compose \
	-f $@.yml \
	-p $@ up \
	--build \
	--renew-anon-volumes \
	--force-recreate

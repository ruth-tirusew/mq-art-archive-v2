# Delegate to monorepo root (run `make test` from mq/ or from here).
ROOT := ../..

.PHONY: help test test-integration migrate api arch-check

help:
	@$(MAKE) -C $(ROOT) help

test test-integration migrate api arch-check:
	@$(MAKE) -C $(ROOT) $@

OUT_DIR := gen/go

.PHONY: generate clean lint

lint:
	docker run --rm \
		-v $(PWD):/defs \
		-w /defs \
		bufbuild/buf \
		lint

generate:
	$(MAKE) clean && docker run --rm \
		-v $(PWD):/defs \
		-w /defs \
		bufbuild/buf \
		generate

clean:
	rm -rf $(OUT_DIR)

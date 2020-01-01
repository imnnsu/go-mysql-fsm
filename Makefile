
path := examples/bin
target := main

.PHONY: example
example:
	@mkdir -p $(path)
	@go build -o $(path)/$(target) examples/main.go
	@$(path)/$(target)

.PHONY: clean
clean:
	rm -rf $(path)

.PHONY: test
test:
	go test
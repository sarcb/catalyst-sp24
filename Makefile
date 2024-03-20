.PHONY: lint
lint:
	golangci-lint run  ./...

.PHONY: fmt
fmt:
	gci write -s standard -s default -s "prefix(github.com/sarcb/catalyst-sp24)" .
	# gofumpt -l -w .
	# wsl --fix ./...
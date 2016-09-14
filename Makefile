.PHONY: all
all:
	gox -osarch="linux/amd64 darwin/amd64 windows/amd64"

.PHONY: vendor
vendor:
	glide install


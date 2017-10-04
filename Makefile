PKG := kkn.fi/heap
NAME := heap

.PHONY: lint
lint:
	gofmt -s -w *.go
	gometalinter ./...

.PHONY: clean
clean:
	@rm -f prof.mem prof.cpu $(NAME).test bench.txt

.PHONY: benchmark
benchmark: clean
	go test -c
	./$(NAME).test -test.bench=. -test.count=5 | tee bench.txt

.PHONY: mem-profile
mem-profile: clean
	go test -run=XXX -bench=. -memprofile=prof.mem $(PKG)
	go tool pprof $(NAME).test prof.mem

.PHONY: cpu-profile
cpu-profile: clean
	go test -run=XXX -bench=. -cpuprofile=prof.cpu $(PKG)
	go tool pprof $(NAME).test prof.cpu

.PHONY: build
build:
	go build $(PKG)

.PHONY: test
test:
	go test $(PKG)

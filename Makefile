
build:
	go build -v ./...

test:
	go test -race -v ./...
watch-test:
	go tool reflex -t 50ms -s -- sh -c 'go tool gotest -race -v ./...'

bench:
	go test -benchmem -count 3 -bench ./...
watch-bench:
	go tool reflex -t 50ms -s -- sh -c 'go test -benchmem -count 3 -bench ./...'

coverage:
	go test -v -coverprofile=cover.out -covermode=atomic ./...
	go tool cover -html=cover.out -o cover.html

tools:
	go get -tool github.com/cespare/reflex@latest
	go get -tool github.com/rakyll/gotest@latest
	go get -tool github.com/psampaz/go-mod-outdated@latest
	go get -tool github.com/jondot/goweight@latest
	go get -tool golang.org/x/tools/cmd/cover
	go get -tool github.com/sonatype-nexus-community/nancy@latest
	go mod tidy

lint:
	golangci-lint run --timeout 60s --max-same-issues 50 ./...
lint-fix:
	golangci-lint run --timeout 60s --max-same-issues 50 --fix ./...

audit:
	go list -json -m all | go tool nancy sleuth

outdated:
	go list -u -m -json all | go tool go-mod-outdated -update -direct

weight:
	go tool goweight

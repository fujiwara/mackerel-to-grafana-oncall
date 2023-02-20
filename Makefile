all: *.go go.*
	go build .

install:
	go install github.com/fujiwara/mackerel-to-grafana-oncall

test:
	go test -v ./...

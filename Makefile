all: *.go go.* cmd/mackerel-to-grafana-oncall/*
	go build -o mackerel-to-grafana-oncall cmd/mackerel-to-grafana-oncall/main.go

install:
	go install github.com/fujiwara/mackerel-to-grafana-oncall/cmd/mackerel-to-grafana-oncall

test:
	go test -v ./...

clean:
	rm -f mackerel-to-grafana-oncall

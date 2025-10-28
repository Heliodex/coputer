module github.com/Heliodex/coputer/wallflower

go 1.25.3

require (
	github.com/Heliodex/coputer/bundle v0.0.0-20250622152943-83f44d21f6b9
	github.com/Heliodex/coputer/litecode v0.0.0-20250622152943-83f44d21f6b9
	github.com/quic-go/quic-go v0.54.1
	github.com/syncthing/notify v0.0.0-20250528144937-c7027d4f7465
	golang.org/x/crypto v0.39.0
)

require (
	go.uber.org/mock v0.5.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
)

replace github.com/Heliodex/coputer/bundle => ../bundle

replace github.com/Heliodex/coputer/litecode => ../litecode

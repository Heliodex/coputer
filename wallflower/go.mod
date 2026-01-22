module github.com/Heliodex/coputer/wallflower

go 1.25.3

require (
	github.com/Heliodex/coputer/bundle v0.0.0-20250622152943-83f44d21f6b9
	github.com/Heliodex/coputer/litecode v0.0.0-20250622152943-83f44d21f6b9
	github.com/quic-go/quic-go v0.57.0
	github.com/syncthing/notify v0.0.0-20250528144937-c7027d4f7465
	golang.org/x/crypto v0.45.0
)

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
)

replace github.com/Heliodex/coputer/bundle => ../bundle

replace github.com/Heliodex/coputer/litecode => ../litecode

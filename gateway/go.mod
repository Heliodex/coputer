module github.com/Heliodex/coputer/gateway

go 1.25.3

require (
	github.com/Heliodex/coputer/litecode v0.0.0-20250622152943-83f44d21f6b9
	github.com/Heliodex/coputer/wallflower v0.0.0-20250707074553-45e33b575a74
)

require (
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace github.com/Heliodex/coputer/litecode => ../litecode

replace github.com/Heliodex/coputer/wallflower => ../wallflower

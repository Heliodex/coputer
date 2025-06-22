module github.com/Heliodex/coputer/wallflower

go 1.24.4

require (
	github.com/Heliodex/coputer/bundle v0.0.0-20250622152943-83f44d21f6b9
	github.com/Heliodex/coputer/litecode v0.0.0-20250622152943-83f44d21f6b9
	golang.org/x/crypto v0.39.0
)

require golang.org/x/sys v0.33.0 // indirect

replace github.com/Heliodex/coputer/litecode => ../litecode

replace github.com/Heliodex/coputer/bundle => ../bundle

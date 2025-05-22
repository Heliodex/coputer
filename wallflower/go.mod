module github.com/Heliodex/coputer/wallflower

go 1.24.3

require (
	github.com/Heliodex/coputer/bundle v0.0.0-20250514224108-01f2477995d4
	github.com/Heliodex/coputer/litecode v0.0.0-20250324181716-ceddb1aa0328
	golang.org/x/crypto v0.35.0
)

require golang.org/x/sys v0.30.0 // indirect

replace github.com/Heliodex/coputer/litecode => ../litecode

replace github.com/Heliodex/coputer/bundle => ../bundle

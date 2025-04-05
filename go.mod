module github.com/Heliodex/coputer

go 1.24.2

require golang.org/x/crypto v0.33.0

require (
	github.com/Heliodex/coputer/litecode v0.0.0-20250324181716-ceddb1aa0328
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace github.com/Heliodex/coputer/litecode => ./litecode

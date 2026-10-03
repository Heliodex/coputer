# Code style guidelines

- Replace em-dashes (—) with en-dashes (–) in text or comments
	- Use em-dashes only as an explicit empty placeholder value
- Replace en-dashes (–) with hyphens (-) in numerical or temporal ranges

## Go

- Use range-over-number `for i := range 5` instead of classic for loops `for i := 0; i < 5; i++` when the index is not mutated inside the loops
	- Same for range-over-slice `for i := range slice` instead of `for i := 0; i < len(slice); i++`
- Use guard clauses where possible to reduce nesting and improve readability
- Put function declarations above their usage in the same file, never below

## Comments

- Do not hard-wrap. Never split a sentence over multiple lines
	- If the new line is after a semicolon or is a new sentence, OK

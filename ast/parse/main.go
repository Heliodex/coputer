package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LuauExt is the extension of the Luau source files the formatter handles.
const LuauExt = ".luau"

// parseErrors renders the parser's errors as one error listing each failure.
func parseErrors(errs []ParseError) error {
	var b strings.Builder

	b.WriteString("parse failed with errors:")

	for _, err := range errs {
		fmt.Fprintf(&b, "\n- %s at %s", err.Message, err.Location)
	}

	return errors.New(b.String())
}

// formatSource formats Luau source, returning its formatted form.
func formatSource(src string) (string, error) {
	ok, res := Parse(src, Options{})
	if !ok {
		return "", parseErrors(res.Errors)
	}

	return formatAST(res).Root.Source(), nil
}

// formatFile reads and formats the .luau file at path, returning its formatted source, which ends with a newline, along with the original content.
func formatFile(path string) (formatted, original string, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("read file: %w", err)
	}

	formatted, err = formatSource(string(content))
	if err != nil {
		return "", "", err
	}

	return formatted + "\n", string(content), nil
}

// writeFormattedFile formats the .luau file at path in place, reporting whether it changed, and leaving a file that already matches untouched so its timestamp and permissions survive.
func writeFormattedFile(path string) (bool, error) {
	formatted, original, err := formatFile(path)
	if err != nil {
		return false, fmt.Errorf("%s: %w", path, err)
	}

	if formatted == original {
		return false, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("%s: get file info: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(formatted), info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("%s: write file: %w", path, err)
	}

	return true, nil
}

// printFormattedFile prints the formatted source of the .luau file at path.
func printFormattedFile(path string) error {
	formatted, _, err := formatFile(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	fmt.Print(formatted)

	return nil
}

// formatDir writes the formatted version of every .luau file inside dir, including those in its subdirectories, carrying on past files that fail so one bad file doesn't hold up the rest. It returns the number of files changed and an error listing every failure.
func formatDir(dir string) (int, error) {
	var (
		changed int
		errs    []error
	)

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		if entry.IsDir() || filepath.Ext(entry.Name()) != LuauExt {
			return nil
		}

		wrote, err := writeFormattedFile(path)
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		if wrote {
			changed++
		}

		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("walk directory: %w", err))
	}

	return changed, errors.Join(errs...)
}

// plural renders a count with its noun, adding an s to the noun for any count but one.
func plural(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, noun)
	}

	return fmt.Sprintf("%d %ss", count, noun)
}

// formatPath formats the .luau file at path, or every .luau file directly inside it when it is a directory.
func formatPath(path string, write bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("get file info: %w", err)
	}

	if info.IsDir() {
		if !write {
			return errors.New("can't print a directory, pass -w to format its files in place")
		}

		formatted, err := formatDir(path)
		if formatted > 0 {
			fmt.Printf("formatted %s in %s\n", plural(formatted, "file"), path)
		}

		return err
	}

	if filepath.Ext(path) != LuauExt {
		return fmt.Errorf("%s is not a %s file", path, LuauExt)
	}

	if write {
		_, err := writeFormattedFile(path)
		return err
	}

	return printFormattedFile(path)
}

// run formats the source read from stdin, or the paths given on the command line.
func run(paths []string, write bool) error {
	if len(paths) > 1 {
		return errors.New("expected at most one file or directory")
	}

	if len(paths) == 0 {
		if write {
			return errors.New("-w needs a file or directory to write to")
		}

		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}

		formatted, err := formatSource(string(content))
		if err != nil {
			return err
		}

		fmt.Println(formatted)

		return nil
	}

	return formatPath(paths[0], write)
}

// usage prints how to use the formatter.
func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-w] [path]\n\nFormat Luau source read from stdin, or the .luau file at path.\nWith -w, write the result back in place instead of printing it. When path is a directory, format every .luau file inside it, in its subdirectories too.\n\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
}

// main formats the Luau source read from stdin and writes it to stdout, or formats the files given with -w in place.
func main() {
	write := flag.Bool("w", false, "write formatted files in place instead of printing them")

	flag.Usage = usage
	flag.Parse()

	if err := run(flag.Args(), *write); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

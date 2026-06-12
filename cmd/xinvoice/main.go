// Command xinvoice converts and validates XRechnung 3.0.2 invoices.
//
// Usage:
//
//	xinvoice convert [--to json|ubl|cii] [--in FILE] [--out FILE]
//	xinvoice validate [--lang de|en] [--in FILE]
//	xinvoice version
//
// Input is read from stdin (or --in) and auto-detected as XML (UBL or CII) or
// JSON (the syntax-neutral model). `validate` exits non-zero when the invoice
// has error-severity findings, so it can gate a pipeline.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	xinvoice "github.com/andeedotnet/go-xinvoice"
)

// version is the module version, kept in sync with the v0.1.0 git tag.
const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "convert":
		os.Exit(cmdConvert(os.Args[2:]))
	case "validate":
		os.Exit(cmdValidate(os.Args[2:]))
	case "-h", "--help", "help":
		usage(os.Stdout)
		os.Exit(0)
	case "version", "-v", "--version":
		fmt.Println("xinvoice", version)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "xinvoice: unknown command %q\n\n", os.Args[1])
		usage(os.Stderr)
		os.Exit(2)
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `xinvoice — convert and validate XRechnung 3.0.2 invoices

Commands:
  convert   Convert between JSON, UBL and CII
            --to json|ubl|cii   target format (default json)
            --in  FILE          input file (default stdin)
            --out FILE          output file (default stdout)

  validate  Validate against the EN16931 / XRechnung rules
            --lang de|en        message language (default en)
            --in  FILE          input file (default stdin)
            exit status 1 when the invoice has errors

  version   Print the xinvoice version

Input is auto-detected as UBL/CII XML or model JSON.
`)
}

func cmdConvert(args []string) int {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	to := fs.String("to", "json", "target format: json|ubl|cii")
	in := fs.String("in", "-", "input file (- for stdin)")
	out := fs.String("out", "-", "output file (- for stdout)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	inv, err := parseInput(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice:", err)
		return 1
	}

	var result []byte
	switch *to {
	case "json":
		result, err = inv.ToJSON()
	case "ubl":
		result, err = inv.ToXML(xinvoice.UBL)
	case "cii":
		result, err = inv.ToXML(xinvoice.CII)
	default:
		fmt.Fprintf(os.Stderr, "xinvoice: unknown --to %q (want json|ubl|cii)\n", *to)
		return 2
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice:", err)
		return 1
	}
	if err := writeOutput(*out, result); err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice:", err)
		return 1
	}
	return 0
}

func cmdValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	lang := fs.String("lang", "en", "message language: de|en")
	in := fs.String("in", "-", "input file (- for stdin)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	inv, err := parseInput(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice:", err)
		return 1
	}
	res := xinvoice.Validate(inv)
	j, err := res.JSON(*lang)
	if err != nil {
		fmt.Fprintln(os.Stderr, "xinvoice:", err)
		return 1
	}
	fmt.Println(string(j))
	if !res.Valid() {
		return 1
	}
	return 0
}

// parseInput reads name (or stdin) and parses XML or JSON into an invoice.
func parseInput(name string) (*xinvoice.Invoice, error) {
	data, err := readInput(name)
	if err != nil {
		return nil, err
	}
	if t := bytes.TrimSpace(data); len(t) > 0 && t[0] == '<' {
		return xinvoice.ParseXML(data)
	}
	return xinvoice.FromJSON(data)
}

func readInput(name string) ([]byte, error) {
	if name == "-" || name == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(name)
}

func writeOutput(name string, data []byte) error {
	if name == "-" || name == "" {
		_, err := os.Stdout.Write(append(data, '\n'))
		return err
	}
	return os.WriteFile(name, data, 0o644)
}

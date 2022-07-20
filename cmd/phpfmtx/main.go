package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"mibk.io/php/ast"
)

var inPlace = flag.Bool("w", false, "write to file")

func main() {
	flag.Parse()
	log.SetPrefix("phpfmt: ")
	log.SetFlags(0)

	if flag.NArg() == 0 {
		if *inPlace {
			log.Fatal("cannot use -w with standard input")
		}
		if err := formatFile("<stdin>", os.Stdout, os.Stdin); err != nil {
			log.Fatal(err)
		}
		return
	}

	for _, filename := range flag.Args() {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		fi, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		perm := fi.Mode().Perm()

		buf := new(bytes.Buffer)
		err = formatFile(filename, buf, f)
		f.Close()
		if err != nil {
			log.Println(err)
			continue
		}

		if *inPlace {
			// TODO: Make backup file?
			if err := ioutil.WriteFile(filename, buf.Bytes(), perm); err != nil {
				log.Fatal(err)
			}
		} else {
			if _, err := io.Copy(os.Stdout, buf); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func formatFile(filename string, out io.Writer, in io.Reader) error {
	file, err := ast.Parse(in)
	if se, ok := err.(*ast.SyntaxError); ok {
		return fmt.Errorf("%s:%d:%d: %v", filename, se.Line, se.Column, se.Err)
	} else if err != nil {
		return err
	}
	return ast.Fprint(out, file)
}

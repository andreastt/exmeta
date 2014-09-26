// Extracts top meta data section in the given source document
// literally.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

type metadata map[string]*json.RawMessage

const eof = -1

type parser struct {
	input *bufio.Reader
}

func (p *parser) next() rune {
	r, _, err := p.input.ReadRune()
	if err == io.EOF {
		return eof
	}
	return r
}

func (p *parser) peek() (rune, error) {
	lead, err := p.input.Peek(1)
	if err == io.EOF {
		return eof, nil
	}

	b, err := p.input.Peek(runeLen(lead[0]))
	if err == io.EOF {
		return eof, nil
	}

	r, _ := utf8.DecodeRune(b)
	return r, nil
}

func runeLen(lead byte) int {
	if lead < 0xC0 {
		return 1
	} else if lead < 0xE0 {
		return 2
	} else if lead < 0xF0 {
		return 3
	} else {
		return 4
	}
}

func extract(input io.Reader) []byte {
	var rv []byte
	p := &parser{input: bufio.NewReader(input)}

	first, _ := p.peek()
	if first == '{' {
		insideMeta, insideLiteral := true, false
		for {
			r, err := p.peek()
			if r == eof || err != nil {
				break
			}

			r = p.next()
			rv = append(rv, byte(r))
			if insideMeta && r == '}' {
				break
			}
			insideLiteral = !insideLiteral && r == '"'
		}
	}

	return rv
}

var (
	raw  bool
	help bool
)

func exit(msg ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: error: %s", os.Args[0], fmt.Sprintln(msg...))
	os.Exit(1)
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-r] DOCUMENT\n", os.Args[0])
	}
	flag.BoolVar(&raw, "r", false, "")
	flag.BoolVar(&help, "h", false, "")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(127)
	}
	if flag.NArg() != 1 {
		exit("missing source document")
	}
	doc := flag.Arg(0)

	fh, err := os.Open(doc)
	if err != nil {
		exit(err)
	}
	data := extract(fh)

	if raw {
		fmt.Println(string(data))
	} else {
		var meta metadata
		err = json.Unmarshal(data, &meta)
		if err != nil {
			exit(err)
		}
		out, err := json.MarshalIndent(meta, "", "  ")
		if err != nil {
			exit(err)
		}
		fmt.Println(string(out))
	}
}

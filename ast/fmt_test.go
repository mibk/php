package ast_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mibk/diff"
	"mibk.io/php/ast"
)

func TestFormatting(t *testing.T) {
	files, err := filepath.Glob("testdata/*.input")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".input")
		t.Run(name, func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatal(err)
			}
			pf, err := ast.Parse(f)
			f.Close()
			if err != nil {
				t.Fatal(err)
			}

			buf := new(bytes.Buffer)
			if err := ast.Fprint(buf, pf); err != nil {
				t.Fatal(err)
			}

			want, err := ioutil.ReadFile(filepath.Join("testdata", name+".golden"))
			if err != nil {
				t.Fatal(err)
			}
			if diff := diffLines(buf.Bytes(), want); diff != "" {
				t.Errorf("files don't match (-got +want)\n%s", diff)
			}
		})
	}
}

func diffLines(a, b []byte) string {
	linesA := bytes.Split(a, []byte("\n"))
	linesB := bytes.Split(b, []byte("\n"))
	return diffByteSlices(linesA, linesB, slices{linesA, linesB})
}

func diffByteSlices(a, b [][]byte, data diff.Data) string {
	eds := diff.Diff(data)
	if len(eds) == 0 {
		return ""
	}

	eds = append(eds, diff.Edit{Index: len(a), Op: diff.None})
	var i int
	var buf strings.Builder
	for _, ed := range eds {
		for ; i < ed.Index; i++ {
			fmt.Fprintf(&buf, " %s\n", a[i])
		}
		if ed.Op == diff.Delete {
			fmt.Fprintf(&buf, "-%s\n", a[i])
			i++
		} else if ed.Op == diff.Insert {
			fmt.Fprintf(&buf, "+%s\n", b[ed.Bindex])
		}
	}
	return buf.String()
}

type slices struct {
	a, b [][]byte
}

func (d slices) Lens() (n, m int) { return len(d.a), len(d.b) }
func (d slices) Equal(i, j int) bool {
	return bytes.Equal(d.a[i], d.b[j])
}

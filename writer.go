package graphql

import (
	"bytes"
	"fmt"
)

type Writer struct {
	bytes.Buffer
	level int
}

func (w *Writer) Indent() {
	for i := 0; i < w.level; i++ {
		w.WriteString("  ")
	}
}

func (w *Writer) Scope(label string, f func()) {
	if len(label) > 0 {
		w.Println(label, "{")
	} else {
		w.Println("{")
	}
	w.level++
	f()
	w.level--
	w.Println("}")
}

func (w *Writer) Println(a ...interface{}) {
	w.Indent()
	fmt.Fprintln(&w.Buffer, a...)
}

package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	p := NewParser("<input>")
	ls := NewReaderLineSource(os.Stdin, func() error {
		_, err := fmt.Printf("> ")
		return err
	})
	for {
		err := p.ParseNext(ls)
		if err != nil {
			if err == io.EOF {
				break
			}
			if IsHandledError(err) {
				_, err = fmt.Println(err)
				if err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
	}
}

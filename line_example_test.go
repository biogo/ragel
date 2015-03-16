// Copyright ©2013 The bíogo.ragel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ragel_test

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/biogo/ragel"
)

var r = strings.NewReader("Hello,\nWorld!\n")

func Example_1() {
	var (
		p, pe, eof int
		data       = make([]byte, 1<<20)
	)

	br, err := ragel.NewBlockReader(r, &p, &pe, &eof, data)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		_, err = br.Read()
		if err != nil {
			fmt.Println(err)
			break
		}
		err = br.BackupTo('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		// %% write exec;
	}
}

func Example_2() {
	var (
		p, pe, eof int
		data       []byte
	)

	ar, err := ragel.NewAppendReader(bufio.NewReader(r), &p, &pe, &eof, &data, '\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		_, err = ar.Read()
		if err != nil {
			fmt.Println(err)
			break
		}

		// %% write exec;
	}
}

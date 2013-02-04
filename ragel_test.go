// Copyright ©2013 The bíogo.ragel Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ragel

import (
	"bufio"
	"io"
	check "launchpad.net/gocheck"
	"strings"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

type res struct {
	data    string
	readErr error
	backErr error
}

func (s *S) TestBlock(c *check.C) {
	for i, t := range []struct {
		in     string
		size   int
		backup byte
		fn     func(p, pe *int)
		out    []res
	}{
		{
			in:   "Hello,\nWorld!\n",
			size: 10,
			fn:   func(p, pe *int) { *p = *pe },
			out: []res{
				{"Hello,\nWor", nil, nil},
				{"ld!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
		{
			in:   "Hello,\nWorld!\n",
			size: 10,
			fn:   func(p, pe *int) {},
			out: []res{
				{"Hello,\nWor", nil, nil},
				{"Hello,\nWor", ErrBufferFull, nil},
			},
		},
		{
			in:     "Hello,\nWorld!\n",
			size:   10,
			backup: '\n',
			fn:     func(p, pe *int) { *p = *pe },
			out: []res{
				{"Hello,\n", nil, nil},
				{"World!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
		{
			in:     "Hello, World!\n",
			size:   10,
			backup: '\n',
			fn:     func(p, pe *int) { *p = *pe },
			out: []res{
				{"Hello, Wor", nil, ErrNotFound},
				{"ld!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
	} {
		r := strings.NewReader(t.in)
		var (
			p, pe, eof int
			data       = make([]byte, t.size)
		)
		br, err := NewBlockReader(r, &p, &pe, &eof, data)
		c.Assert(err, check.Equals, nil)

		for k, e := range t.out {
			_, err = br.Read()
			c.Check(err, check.Equals, e.readErr, check.Commentf("Test %d, Line %d", i, k))
			if t.backup != 0 && err == nil {
				err = br.BackupTo(t.backup)
				c.Check(err, check.Equals, e.backErr, check.Commentf("Test %d, Line %d", i, k))
			}
			c.Check(string(data[p:pe]), check.Equals, e.data, check.Commentf("Test %d, Line %d", i, k))
			t.fn(&p, &pe)
		}
	}
}

func (s *S) TestAppend(c *check.C) {
	for i, t := range []struct {
		in     string
		delim  byte
		backup byte
		fn     func(p, pe *int)
		out    []res
	}{
		{
			in:    "Hello,\nWorld!\n",
			delim: '\n',
			fn:    func(p, pe *int) { *p = *pe },
			out: []res{
				{"Hello,\n", nil, nil},
				{"World!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
		{
			in:    "Hello,\nWorld!\n",
			delim: '\n',
			fn:    func(p, pe *int) {},
			out: []res{
				{"Hello,\n", nil, nil},
				{"Hello,\nWorld!\n", nil, nil},
				{"Hello,\nWorld!\n", io.EOF, nil},
			},
		},
		{
			in:    "Hello,\nWorld!",
			delim: '\n',
			fn:    func(p, pe *int) {},
			out: []res{
				{"Hello,\n", nil, nil},
				{"Hello,\nWorld!", io.ErrUnexpectedEOF, nil},
			},
		},
		{
			in:     "Hello,\nWorld!\n",
			delim:  '\n',
			backup: '\n',
			fn:     func(p, pe *int) { *p = *pe },
			out: []res{
				{"Hello,\n", nil, nil},
				{"World!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
	} {
		r := strings.NewReader(t.in)
		var (
			p, pe, eof int
			data       []byte
		)
		br, err := NewAppendReader(bufio.NewReader(r), &p, &pe, &eof, &data, t.delim)
		c.Assert(err, check.Equals, nil)

		for k, e := range t.out {
			_, err = br.Read()
			c.Check(err, check.Equals, e.readErr, check.Commentf("Test %d, Line %d", i, k))
			if t.backup != 0 && err == nil {
				err = br.BackupTo(t.backup)
				c.Check(err, check.Equals, e.backErr, check.Commentf("Test %d, Line %d", i, k))
			}
			c.Check(string(data[p:pe]), check.Equals, e.data, check.Commentf("Test %d, Line %d", i, k))
			t.fn(&p, &pe)
		}
	}
}

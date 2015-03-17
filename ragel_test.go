// Copyright ©2013 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ragel

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

// 'Line' reader support tests

type blockRes struct {
	data    string
	readErr error
	backErr error
}

func (s *S) TestBlockReader(c *check.C) {
	for i, t := range []struct {
		in     string
		size   int
		backup byte
		fn     func(p, pe *int)
		out    []blockRes
	}{
		{
			in:   "Hello,\nWorld!\n",
			size: 10,
			fn:   func(p, pe *int) { *p = *pe },
			out: []blockRes{
				{"Hello,\nWor", nil, nil},
				{"ld!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
		{
			in:   "Hello,\nWorld!\n",
			size: 10,
			fn:   func(p, pe *int) {},
			out: []blockRes{
				{"Hello,\nWor", nil, nil},
				{"Hello,\nWor", ErrBufferFull, nil},
			},
		},
		{
			in:     "Hello,\nWorld!\n",
			size:   10,
			backup: '\n',
			fn:     func(p, pe *int) { *p = *pe },
			out: []blockRes{
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
			out: []blockRes{
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

func (s *S) TestAppendReader(c *check.C) {
	for i, t := range []struct {
		in     string
		delim  byte
		backup byte
		fn     func(p, pe *int)
		out    []blockRes
	}{
		{
			in:    "Hello,\nWorld!\n",
			delim: '\n',
			fn:    func(p, pe *int) { *p = *pe },
			out: []blockRes{
				{"Hello,\n", nil, nil},
				{"World!\n", nil, nil},
				{"", io.EOF, nil},
			},
		},
		{
			in:    "Hello,\nWorld!\n",
			delim: '\n',
			fn:    func(p, pe *int) {},
			out: []blockRes{
				{"Hello,\n", nil, nil},
				{"Hello,\nWorld!\n", nil, nil},
				{"Hello,\nWorld!\n", io.EOF, nil},
			},
		},
		{
			in:    "Hello,\nWorld!",
			delim: '\n',
			fn:    func(p, pe *int) {},
			out: []blockRes{
				{"Hello,\n", nil, nil},
				{"Hello,\nWorld!", io.ErrUnexpectedEOF, nil},
			},
		},
		{
			in:     "Hello,\nWorld!\n",
			delim:  '\n',
			backup: '\n',
			fn:     func(p, pe *int) { *p = *pe },
			out: []blockRes{
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

// Scanner support tests

type scanRes struct {
	data    string
	readErr error
	token   string
}

func quoteScan(p, pe, ts, te *int, data string) {
	if data[*ts] != '"' {
		*ts = strings.Index(data, `"`)
	} else {
		*ts = strings.Index(data[1:], `"`) + 1
	}
	if *ts < 0 {
		*ts = 0
	}
	*p = *pe
}

func (s *S) TestBlockScanner(c *check.C) {
	for i, t := range []struct {
		in   string
		size int
		fn   func(p, pe, ts, te *int, data string)
		out  []scanRes
	}{
		{
			in:   `A stream "of characters" to be scanned. More to come.`,
			size: 20,
			fn:   func(p, pe, ts, te *int, data string) { *p, *ts = *pe, *pe },
			out: []scanRes{
				{`A stream "of charact`, nil, ""},
				{`ers" to be scanned. `, nil, ""},
				{`More to come.`, nil, ""},
				{"", io.EOF, ""},
			},
		},
		{
			in:   `A stream "of characters" to be scanned. More to come.`,
			size: 10,
			fn:   func(p, pe, ts, te *int, data string) {},
			out: []scanRes{
				{`A stream "`, nil, ""},
				{`A stream "`, ErrBufferFull, ""},
			},
		},
		{
			in:   `A stream "of characters" to be scanned. More to come.`,
			size: 20,
			fn: func(p, pe, ts, te *int, data string) {
				*p, *ts = *pe, *pe
			},
			out: []scanRes{
				{`A stream "of charact`, nil, ""},
				{`ers" to be scanned. `, nil, ""},
				{`More to come.`, nil, ""},
				{"", io.EOF, ""},
			},
		},
		{
			in:   `A stream "of characters" to be scanned. More to come.`,
			size: 15,
			fn:   quoteScan,
			out: []scanRes{
				{`A stream "of ch`, nil, ""},
				{`aracters"`, nil, `"of characters"`},
				{` to be scanned`, nil, `" to be scanned`},
				{"", ErrBufferFull, `" to be scanned`},
			},
		},
		{
			in:   `A stream "of characters" to be scanned. "More" to come.`,
			size: 30,
			fn:   quoteScan,
			out: []scanRes{
				{`A stream "of characters" to be`, nil, ""},
				{` scanned.`, nil, `"of characters" to be scanned.`},
				{` "More" to com`, nil, `" to be scanned. "More" to com`},
				{`e.`, nil, `"More" to come.`},
				{"", io.EOF, `" to come.`},
			},
		},
	} {
		r := strings.NewReader(t.in)
		var (
			p, pe, eof int
			ts, te     int
			data       = make([]byte, t.size)
		)
		br, err := NewBlockScanner(r, &p, &pe, &ts, &te, &eof, data)
		c.Assert(err, check.Equals, nil)

		for k, e := range t.out {
			_, err = br.Read()
			c.Check(err, check.Equals, e.readErr, check.Commentf("Test %d, Line %d", i, k))
			c.Check(string(data[p:pe]), check.Equals, e.data, check.Commentf("Test %d, Line %d", i, k))
			if ts < p {
				c.Check(string(data[ts:pe]), check.Equals, e.token, check.Commentf("Test %d, Line %d", i, k))
			}
			t.fn(&p, &pe, &ts, &te, string(data))
		}
	}
}

func (s *S) TestAppendScanner(c *check.C) {
	for i, t := range []struct {
		in    string
		delim byte
		fn    func(p, pe, ts, te *int, data string)
		out   []scanRes
	}{
		{
			in:    `A stream "of characters" to be scanned. More to come.`,
			delim: '.',
			fn:    func(p, pe, ts, te *int, data string) { *p, *ts = *pe, *pe },
			out: []scanRes{
				{`A stream "of characters" to be scanned.`, nil, ""},
				{` More to come.`, nil, ""},
				{"", io.EOF, ""},
			},
		},
		{
			in:    `A stream "of characters" to be scanned. More to come.`,
			delim: '.',
			fn:    func(p, pe, ts, te *int, data string) {},
			out: []scanRes{
				{`A stream "of characters" to be scanned.`, nil, ""},
				{`A stream "of characters" to be scanned. More to come.`, nil, ""},
				{`A stream "of characters" to be scanned. More to come.`, io.EOF, ""},
			},
		},
		{
			in:    `A stream "of characters" to be scanned. More to come`,
			delim: '.',
			fn:    func(p, pe, ts, te *int, data string) {},
			out: []scanRes{
				{`A stream "of characters" to be scanned.`, nil, ""},
				{`A stream "of characters" to be scanned. More to come`, io.ErrUnexpectedEOF, ""},
			},
		},
		{
			in:    `A stream "of characters" to be scanned. More to come.`,
			delim: '.',
			fn:    quoteScan,
			out: []scanRes{
				{`A stream "of characters" to be scanned.`, nil, ""},
				{` More to come.`, nil, `"of characters" to be scanned. More to come.`},
				{"", io.EOF, `" to be scanned. More to come.`},
			},
		},
	} {
		r := strings.NewReader(t.in)
		var (
			p, pe, eof int
			ts, te     int
			data       []byte
		)
		br, err := NewAppendScanner(bufio.NewReader(r), &p, &pe, &ts, &te, &eof, &data, t.delim)
		c.Assert(err, check.Equals, nil)

		for k, e := range t.out {
			_, err = br.Read()
			c.Check(err, check.Equals, e.readErr, check.Commentf("Test %d, Line %d", i, k))
			c.Check(string(data[p:pe]), check.Equals, e.data, check.Commentf("Test %d, Line %d", i, k))
			if ts < p {
				c.Check(string(data[ts:]), check.Equals, e.token, check.Commentf("Test %d, Line %d", i, k))
			}
			t.fn(&p, &pe, &ts, &te, string(data))
		}
	}
}

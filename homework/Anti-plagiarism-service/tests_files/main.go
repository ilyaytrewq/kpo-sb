package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Node struct {
	val   int64
	sum   int64
	sz    int
	prior uint32
	l, r  *Node
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func newNode(x int64) *Node {
	return &Node{
		val:   x,
		sum:   x,
		sz:    1,
		prior: rng.Uint32(),
	}
}

func size(v *Node) int {
	if v == nil {
		return 0
	}
	return v.sz
}

func sm(v *Node) int64 {
	if v == nil {
		return 0
	}
	return v.sum
}

func upd(v *Node) {
	if v == nil {
		return
	}
	v.sz = 1 + size(v.l) + size(v.r)
	v.sum = v.val + sm(v.l) + sm(v.r)
}

func merge(a, b *Node) *Node {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.prior > b.prior {
		a.r = merge(a.r, b)
		upd(a)
		return a
	}
	b.l = merge(a, b.l)
	upd(b)
	return b
}

// split(v, k) -> (A, B) where A has first k elements (by implicit index), B has the rest
func split(v *Node, k int) (*Node, *Node) {
	if v == nil {
		return nil, nil
	}
	if 1+size(v.l) <= k {
		a, b := split(v.r, k-1-size(v.l))
		v.r = a
		upd(v)
		return v, b
	}
	a, b := split(v.l, k)
	v.l = b
	upd(v)
	return a, v
}

type FastScanner struct {
	r *bufio.Reader
}

func NewFastScanner() *FastScanner {
	return &FastScanner{r: bufio.NewReaderSize(os.Stdin, 1<<20)}
}

func (fs *FastScanner) nextInt() int {
	sign := 1
	c, err := fs.r.ReadByte()
	for err == nil && (c == ' ' || c == '\n' || c == '\r' || c == '\t') {
		c, err = fs.r.ReadByte()
	}
	if err != nil {
		return 0
	}
	if c == '-' {
		sign = -1
		c, _ = fs.r.ReadByte()
	}
	x := 0
	for c >= '0' && c <= '9' {
		x = x*10 + int(c-'0')
		c, err = fs.r.ReadByte()
		if err != nil {
			break
		}
	}
	if err == nil {
		_ = fs.r.UnreadByte()
	}
	return x * sign
}

func main() {
	in := NewFastScanner()
	out := bufio.NewWriterSize(os.Stdout, 1<<20)
	defer out.Flush()

	var even, odd *Node
	iters := 1

	for {
		n := in.nextInt()
		q := in.nextInt()
		if n+q == 0 {
			return
		}

		fmt.Fprintf(out, "Swapper %d:\n", iters)
		iters++

		even, odd = nil, nil

		for i := 0; i < n; i++ {
			x := int64(in.nextInt())
			if i&1 == 1 {
				odd = merge(odd, newNode(x))
			} else {
				even = merge(even, newNode(x))
			}
		}

		for ; q > 0; q-- {
			typ := in.nextInt()
			l := in.nextInt()
			r := in.nextInt()
			l--
			r--
			length := r - l + 1

			if typ == 1 {
				q1l := l / 2
				q1r := q1l + length/2
				q2l := (l + 1) / 2
				q2r := q2l + length/2
				if l&1 == 1 {
					q1l, q2l = q2l, q1l
					q1r, q2r = q2r, q1r
				}

				a, b := split(even, q1r)
				c, d := split(a, q1l)

				a2, b2 := split(odd, q2r)
				c2, d2 := split(a2, q2l)

				even = merge(c, merge(d2, b))
				odd = merge(c2, merge(d, b2))
			} else {
				q1l := l / 2
				q1r := q1l + (length+1)/2
				q2l := (l + 1) / 2
				q2r := q2l + length/2
				if l&1 == 1 {
					q1l, q2l = q2l, q1l
					q1r, q2r = q2r, q1r
				} //
				////
				a, b := split(even, q1r)
				c, d := split(a, q1l)

				a2, b2 := split(odd, q2r)
				c2, d2 := split(a2, q2l)

				fmt.Fprintf(out, "%d\n", sm(d)+sm(d2))

				even = merge(c, merge(d, b))
				odd = merge(c2, merge(d2, b2))
			}
		}

		fmt.Fprintln(out)
	}
}

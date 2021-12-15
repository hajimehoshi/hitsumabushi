// SPDX-License-Identifier: Apache-2.0

package hitsumabushi

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type patch struct {
	name     string
	replaces map[string]string
	append   string
}

func parsePatch(name string, r io.Reader) (*patch, error) {
	p := &patch{
		name:     name,
		replaces: map[string]string{},
	}
	var from string
	var to string

	const (
		phaseInit = iota
		phaseFrom
		phaseTo
		phaseAppend
	)
	var phase int

	s := bufio.NewScanner(r)
	var i int
	for s.Scan() {
		switch line := s.Text(); line {
		case "//--from":
			if phase == phaseFrom {
				return nil, fmt.Errorf("unexpected //--from at L%d", i)
			}
			if phase != phaseInit {
				p.replaces[from] = to
				from = ""
				to = ""
			}
			phase = phaseFrom

		case "//--to":
			if phase != phaseFrom {
				return nil, fmt.Errorf("unexpected //--to at L%d", i)
			}
			phase = phaseTo

		case "//--append":
			if phase == phaseFrom {
				return nil, fmt.Errorf("unexpected //--append at L%d", i)
			}
			if phase != phaseInit {
				p.replaces[from] = to
				from = ""
				to = ""
			}
			phase = phaseAppend

		default:
			switch phase {
			case phaseInit:
				return nil, fmt.Errorf("unexpected content at L%d", i)
			case phaseFrom:
				from += line + "\n"
			case phaseTo:
				to += line + "\n"
			case phaseAppend:
				p.append += line + "\n"
			}

		}
		i++
	}

	if from != "" {
		p.replaces[from] = to
	}
	return p, nil
}

func (p *patch) apply(r io.Reader) (io.Reader, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	str := string(buf)
	for from, to := range p.replaces {
		if !strings.Contains(str, from) {
			return nil, fmt.Errorf("hitsumabushi: patching %s failed: %s", p.name, from[:strings.IndexByte(from, '\n')])
		}
		old := str
		new := strings.Replace(old, from, to, 1)
		if old == new {
			return nil, fmt.Errorf("hitsumabushi: patching %s failed: replacing result is the same", p.name)
		}
		str = new
	}
	str += p.append
	return bytes.NewBufferString(str), nil
}

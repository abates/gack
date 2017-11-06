package gack

import (
	"bytes"
	"strings"
	"unicode"
)

// Match returned when scanning a Pattern
type Match struct {
	subject  string
	match    bool
	captures map[string]string
}

// Subject returns the subject string for this match
func (m *Match) Subject() string {
	return m.subject
}

// Matches indicates whether or not the Pattern was a match
func (m *Match) Matches() bool {
	return m.match
}

// Param will return the named parameter from the matched string
func (m *Match) Param(name string) string {
	return m.captures[name]
}

// Interpolate a subject string with the captured values from a pattern
func (m *Match) Interpolate(subject string) string {
	var buffer bytes.Buffer
	p := NewPattern(subject)
	for _, token := range p.tokens {
		if strings.HasPrefix(token, ":") {
			token = token[1:]
			str := "(MISSING)"
			if s, found := m.captures[token]; found {
				str = s
			}
			buffer.WriteString(str)
		} else {
			buffer.WriteString(token)
		}
	}
	return buffer.String()
}

// Pattern to match for a target. Patterns are strings optionally
// containing WILDCARD characters in positions where the value is
// not known ahead of time
type Pattern struct {
	pattern string
	tokens  []string
}

// NewPattern creates a new pattern from a pattern string.  Patterns
// are simple  expressions similar to glob expressions.  Patterns
// containing the WILDCARD character will consume characters on either
// side of wildcard and capture any characters representing the wildcard.
//
// Captured strings can be used for interpolation with other patterns
// For instance, a pattern such as:
//
//   NewPattern("pkg/*.deb")
//
// Will capture anything between "pkg/" and ".deb".
//
// An example of Using the captured strings from this pattern
// (available in the Match object):
//   package "main"
//
//   import "github.com/abates/gack"
//
//   func main() {
//     p := gack.NewPattern("pkg/*.deb")
//     match := p.Match("pkg/mypackage.deb")
//     str := match.Interpolate("build/*")
//     println(str) // "build/mypackage"
//   }
//
func NewPattern(pattern string) Pattern {
	p := Pattern{
		pattern: pattern,
	}
	l := newLexer(pattern, p.readTextState)
	l.parse()
	p.tokens = l.tokens
	return p
}

func (p *Pattern) finishedState(l *lexer) stateFn {
	l.appendToken()
	return nil
}

func (p *Pattern) readCaptureState(l *lexer) stateFn {
	r := l.next()
	if r == eof {
		return p.finishedState
	} else if unicode.IsLetter(r) {
		return p.readCaptureState
	}
	l.backup()
	l.appendToken()
	return p.readTextState
}

func (p *Pattern) readTextState(l *lexer) stateFn {
	r := l.next()
	if r == eof {
		return p.finishedState
	} else if r == ':' {
		l.backup()
		l.appendToken()
		l.next()
		return p.readCaptureState
	}
	return p.readTextState
}

// Match a subject string against the Pattern
func (p *Pattern) Match(subject string) Match {
	var match Match
	match.subject = subject
	match.captures = make(map[string]string)
	match.match = true
	for i, token := range p.tokens {
		if strings.HasPrefix(token, ":") {
			// remove the colon from the name
			token = token[1:]
			if i == len(p.tokens)-1 {
				match.captures[token] = subject
				break
			} else if index := strings.Index(subject, p.tokens[i+1]); index > -1 {
				match.captures[token] = subject[0:index]
				subject = subject[index:]
			} else {
				match.match = false
				break
			}
		} else if strings.HasPrefix(subject, token) {
			subject = subject[len(token):]
		} else {
			match.match = false
			break
		}
	}

	if match.match {
		match.captures["*"] = match.subject
	}
	return match
}

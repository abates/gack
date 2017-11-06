package gack

import (
	"unicode/utf8"
)

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input      string
	pos        int
	width      int
	tokenStart int
	tokens     []string
	startState stateFn
}

func newLexer(input string, startState stateFn) *lexer {
	return &lexer{
		input:      input,
		startState: startState,
	}
}

func (l *lexer) parse() {
	for state := l.startState; state != nil; {
		state = state(l)
	}
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) appendToken() {
	if l.pos > l.tokenStart {
		l.tokens = append(l.tokens, l.input[l.tokenStart:l.pos])
	}
	l.tokenStart = l.pos
}

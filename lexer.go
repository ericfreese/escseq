package escseq

import (
	"bufio"
	"bytes"
	"io"
)

type lexerState uint8

const (
	lexDefault lexerState = iota
	lexEsc
	lexCSParam
	lexCSInter
)

type Lexer interface {
	ReadToken() (Token, error)
}

type lexer struct {
	lastState lexerState
	state     lexerState
	rd        io.RuneScanner
}

// Returns a Lexer that can read tokens from the provided io.Reader
func NewLexer(rd io.Reader) Lexer {
	return &lexer{lexDefault, lexDefault, bufio.NewReader(rd)}
}

// ReadToken reads a single Token from the input stream. It will
// return the Token and any errors it encountered while reading the
// Token. If no Token can be read, it will return a Token with type
// TokNone. If an unexpected input is found, it will return a Token
// with type TokUnknown.
func (l *lexer) ReadToken() (Token, error) {
	r, _, err := l.rd.ReadRune()

	if err != nil {
		return &token{TokNone, ""}, err
	}

	switch l.state {
	case lexDefault:
		if r == 0x1b {
			l.setState(lexEsc)
			return &token{TokEsc, string(r)}, nil
		} else {
			return &token{TokText, string(r)}, nil
		}
	case lexEsc:
		if r >= 0x40 && r <= 0x5f {
			if r == '[' {
				l.setState(lexCSParam)
			} else {
				l.setState(lexDefault)
			}

			return &token{TokFe, string(r)}, nil
		}
	case lexCSParam:
		switch {
		case r >= 0x40 && r <= 0x7e:
			l.setState(lexDefault)
			return &token{TokFinal, string(r)}, nil
		case l.lastState == lexEsc && r >= 0x3c && r <= 0x3f:
			l.rd.UnreadRune()
			return l.readCSPrivParam()
		case r >= 0x30 && r <= 0x3f:
			l.setState(lexCSParam)

			switch {
			case r >= 0x3c && r <= 0x3f:
				return &token{TokUnknown, string(r)}, nil
			case r == ':':
				return &token{TokParamSep, string(r)}, nil
			case r == ';':
				return &token{TokSep, string(r)}, nil
			default:
				l.rd.UnreadRune()
				return l.readCSParamNum()
			}
		case r >= 0x20 && r <= 0x2f:
			l.setState(lexCSInter)
			return &token{TokInter, string(r)}, nil
		}
	case lexCSInter:
		switch {
		case r >= 0x40 && r <= 0x7e:
			l.setState(lexDefault)
			return &token{TokFinal, string(r)}, nil
		case r >= 0x20 && r <= 0x2f:
			l.setState(lexCSInter)
			return &token{TokInter, string(r)}, nil
		}
	}

	l.setState(lexDefault)
	return &token{TokUnknown, string(r)}, nil
}

func (l *lexer) setState(s lexerState) {
	l.lastState = l.state
	l.state = s
}

func (l *lexer) readCSParamNum() (Token, error) {
	var (
		buf bytes.Buffer
		r   rune
		err error
	)

	for {
		r, _, err = l.rd.ReadRune()

		if err != nil {
			break
		}

		if r >= 0x30 && r <= 0x39 {
			buf.WriteRune(r)
		} else {
			l.rd.UnreadRune()
			break
		}
	}

	return &token{TokParamNum, buf.String()}, err
}

func (l *lexer) readCSPrivParam() (Token, error) {
	var (
		buf bytes.Buffer
		r   rune
		err error
	)

	for {
		r, _, err = l.rd.ReadRune()

		if err != nil {
			break
		}

		if r >= 0x30 && r <= 0x3f {
			buf.WriteRune(r)
		} else {
			l.rd.UnreadRune()
			break
		}
	}

	return &token{TokPrivParam, buf.String()}, err
}

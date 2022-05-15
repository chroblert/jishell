// Package shlex provides a simple lexical analysis like Unix shell.
package shlex

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

var (
	ErrNoClosing = errors.New("no closing quotation")
	ErrNoEscaped = errors.New("no escaped character")
)

// Tokenizer is the interface that classifies a token according to
// words, whitespaces, quotations, escapes and escaped quotations.
type Tokenizer interface {
	IsWord(rune) bool
	IsDelimiter(rune, rune, bool) bool
	IsQuote(rune) bool
	IsEscape(rune) bool
	IsEscapedQuote(rune, bool) bool
}

// DefaultTokenizer implements a simple tokenizer like Unix shell.
type DefaultTokenizer struct{}

func (t *DefaultTokenizer) IsWord(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
func (t *DefaultTokenizer) IsQuote(r rune) bool {
	switch r {
	case '\'', '"':
		return true
	default:
		return false
	}
}
func (t *DefaultTokenizer) IsDelimiter(r rune, delimiter rune, delimiterSpecific bool) bool {
	//return unicode.IsSpace(r)
	//JC 220513
	//jlog.Error("isWhitespace:",r,r==delimiter)
	if delimiterSpecific {
		return r == delimiter
	}
	return unicode.IsSpace(r)
}
func (t *DefaultTokenizer) IsEscape(r rune) bool {
	return r == '\\'
}

// delimiterSpecific 是否指定了分隔符，若未指定，则按照标准的posix模式
func (t *DefaultTokenizer) IsEscapedQuote(r rune, preserveLiteral bool) bool {
	if preserveLiteral {
		return r == '"' || r == '\''
	} else {
		return r == '"'
	}
}

// Lexer represents a lexical analyzer.
type Lexer struct {
	reader            *bufio.Reader
	tokenizer         Tokenizer
	posix             bool
	whitespaceSplit   bool
	delimiter         rune // JC 220514: 指定分隔符
	preserveLiteral   bool // JC 220514: 是否保留所有字面量 如：\,',"
	delimiterSpecific bool // JC 220515: 是否指定分隔符
}

// NewLexer creates a new Lexer reading from io.Reader.  This Lexer
// has a DefaultTokenizer according to posix and whitespaceSplit
// rules.
func NewLexer(posix, whitespaceSplit bool, delimiter rune, r io.Reader, preserveLiteral bool, delimiterSpecific bool) *Lexer {
	return &Lexer{
		reader:            bufio.NewReader(r),
		tokenizer:         &DefaultTokenizer{},
		posix:             posix,
		whitespaceSplit:   whitespaceSplit,
		delimiter:         delimiter,
		preserveLiteral:   preserveLiteral,
		delimiterSpecific: delimiterSpecific,
	}
}

// NewLexerString creates a new Lexer reading from a string.  This
// Lexer has a DefaultTokenizer according to posix and whitespaceSplit
// rules.
func NewLexerString(posix, whitespaceSplit bool, delimiter rune, s string, preserveLiteral, delimiterSpecific bool) *Lexer {
	return NewLexer(posix, whitespaceSplit, delimiter, strings.NewReader(s), preserveLiteral, delimiterSpecific)
}

// Split splits a string according to posix or non-posix rules.
// 默认分隔符为空白(不只是空格),否则为填写的第一个rune
func Split(s string, posix bool, preserveLiteral bool, delimiter ...rune) ([]string, error) {
	// JC 220514: 默认分隔符为空格
	if len(delimiter) == 0 {
		return NewLexerString(posix, true, ' ', s, preserveLiteral, false).Split()
	}
	return NewLexerString(posix, true, delimiter[0], s, preserveLiteral, true).Split()
}

// Split splits a string according to posix or non-posix rules.
// 未修改的Split()
func OriginSplit(s string, posix bool) ([]string, error) {
	return NewLexerString(posix, true, ' ', s, false, false).Split()
}

// SetTokenizer sets a Tokenizer.
func (l *Lexer) SetTokenizer(t Tokenizer) {
	l.tokenizer = t
}

func (l *Lexer) Split() ([]string, error) {
	result := make([]string, 0)
	for {
		token, err := l.readToken()
		//jlog.Error("token:",string(token),token,token==nil)
		if token != nil {
			result = append(result, string(token))
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (l *Lexer) readToken() (token []rune, err error) {
	t := l.tokenizer
	quoted := false
	//state := ' '
	state := l.delimiter
	escapedState := ' '
scanning:
	for {
		next, _, err := l.reader.ReadRune()
		//jlog.Error("state:","_"+string(state)+"_",state)
		//jlog.Error("next:","_"+string(next)+"_",next)
		if err != nil {
			if t.IsQuote(state) {
				return token, ErrNoClosing
			} else if t.IsEscape(state) {
				return token, ErrNoEscaped
			}
			return token, err
		}

		switch {
		case t.IsDelimiter(state, l.delimiter, l.delimiterSpecific):
			switch {
			case t.IsDelimiter(next, l.delimiter, l.delimiterSpecific):
				break scanning
			case l.posix && t.IsEscape(next):
				if l.preserveLiteral {
					token = append(token, next) // JC 220514
				}
				escapedState = 'a'
				state = next
			case t.IsWord(next):
				token = append(token, next)
				state = 'a'
			case t.IsQuote(next):
				if !l.posix {
					token = append(token, next)
				} else {
					if l.preserveLiteral {
						token = append(token, next) // JC 220514
					}
				}
				state = next
			default:
				token = []rune{next}
				if l.whitespaceSplit {
					state = 'a'
				} else if token != nil || (l.posix && quoted) {
					break scanning
				}
			}
		case t.IsQuote(state):
			quoted = true
			switch {
			case next == state:
				if !l.posix {
					token = append(token, next)
					break scanning
				} else {
					if token == nil {
						token = []rune{}
					}
					if l.preserveLiteral {
						token = append(token, next) // JC 220514
					}
					state = 'a'
				}
			case l.posix && t.IsEscape(next) && t.IsEscapedQuote(state, l.preserveLiteral):
				if l.preserveLiteral {
					token = append(token, next) // JC 220514
				}
				escapedState = state
				state = next
			default:
				token = append(token, next)
			}
		case t.IsEscape(state):
			if t.IsQuote(escapedState) && next != state && next != escapedState {
				token = append(token, state)
			}
			token = append(token, next)
			state = escapedState
		case t.IsWord(state):
			switch {
			case t.IsDelimiter(next, l.delimiter, l.delimiterSpecific):
				if token != nil || (l.posix && quoted) {
					break scanning
				}
			case l.posix && t.IsQuote(next):
				if l.preserveLiteral {
					token = append(token, next) // JC 220514
				}
				state = next
			case l.posix && t.IsEscape(next):
				if l.preserveLiteral {
					token = append(token, next) // JC 220514
				}
				escapedState = 'a'
				state = next
			case t.IsWord(next) || t.IsQuote(next):
				token = append(token, next)
			default:
				if l.whitespaceSplit {
					token = append(token, next)
				} else if token != nil {
					l.reader.UnreadRune()
					break scanning
				}
			}
		}
	}
	return token, nil
}

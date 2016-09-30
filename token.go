package escseq

type TokenType uint8

const (
	TokNone TokenType = iota
	TokUnknown
	TokText
	TokEsc
	TokFe
	TokPrivParam
	TokParamNum
	TokParamSep
	TokSep
	TokInter
	TokFinal
)

type Token interface {
	Type() TokenType
	Val() string
}

type token struct {
	t   TokenType
	val string
}

func (t *token) Type() TokenType {
	return t.t
}

func (t *token) Val() string {
	return t.val
}

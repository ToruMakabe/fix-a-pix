package formula

import (
	"strconv"
	"strings"
)

// tseitinConverter はTseitin変換に必要なデータとメソッドをまとめた構造体である.
type tseitinConverter struct {
	fv string
	fc int
}

// newTseitinConverter はtseitinConverterのコンストラクターで, fresh variableの初期化を行う.
func newTseitinConverter() *tseitinConverter {
	return &tseitinConverter{fv: "x", fc: 0}
}

// incFv はtseitinConverterのメソッドで, fresh variableのインデックス部分をインクリメントする.
func (tc *tseitinConverter) incFv() string {
	tc.fc++
	return tc.fv + strconv.Itoa(tc.fc)
}

// conv はtseitinConverterのメソッドで, 構文木を再帰的に副式に分解し, Tseitin変換を行う.
func (tc *tseitinConverter) conv(e /* expression */ Expression, pop /* parent op */ string) ([][]string, string) {

	var (
		rf string
		r  [][]string
	)

	switch e.(type) {
	case BinOpExpr:
		op := string(rune(e.(BinOpExpr).Operator))
		// convの結果, leftに葉から追記を重ねた選言部を, lvに呼び出し先のリテラルか生成したfresh variableを得る. right, rvも同様.
		left, lv := tc.conv(e.(BinOpExpr).Left, op)
		right, rv := tc.conv(e.(BinOpExpr).Right, op)
		// fresh variableを生成する.
		fv := tc.incFv()
		// left,rightをマージする.
		r = append(left, right...)
		if op == "&" {
			// (fersh variable) -> (lv & rv) を選言に変換する.
			return append(r, []string{"~" + fv, lv}, []string{"~" + fv, rv}), fv
		}
		if op == "|" {
			// (fersh variable) -> (lv | rv) を選言に変換する.
			return append(r, []string{"~" + fv, lv, rv}), fv
		}
	case NotOpExpr:
		op := string(rune(e.(NotOpExpr).Operator))
		right, v := tc.conv(e.(NotOpExpr).Right, op)
		return right, v
	case Atomic:
		rf = e.(Atomic).Atomic
		// 親の結合子が否定の場合は否定のリテラルにする.
		if pop == "~" {
			rf = "~" + rf
		}
		return nil, rf
	default:
		return nil, ""
	}
	return nil, ""

}

// ConvNNF は否定標準形への変換を行う.
func ConvNNF(f /* formula */ string) (string, error) {
	var (
		r   *strings.Reader
		p   Expression
		err error
	)

	// 含意を変換するため, goyaccで構文木を作成する.
	r = strings.NewReader(f)
	p, err = Parse(r)
	if err != nil {
		return "", err
	}
	// 含意を変換する. 再帰的に構文木を探索するため, 切り出した別関数を呼び出す.
	fl := convImply(p, "")

	// 否定をアトミックに寄せ, 二重否定を削除するため, goyaccで構文木を作成する.
	r = strings.NewReader(fl)
	p, err = Parse(r)
	if err != nil {
		return "", err
	}
	// 否定をアトミックに寄せ, 二重否定を削除する. 再帰的に構文木を探索するため, 切り出した別関数を呼び出す.
	fl = convNeg(p, "", 0)

	return fl, nil

}

// convImply は構文木を再帰的に探索し,含意を変換する.
func convImply(e /* expression */ Expression, pop /* parent op */ string) string {
	var rf string
	switch e.(type) {
	case BinOpExpr:
		op := string(rune(e.(BinOpExpr).Operator))
		left := convImply(e.(BinOpExpr).Left, op)
		right := convImply(e.(BinOpExpr).Right, op)
		if op == ">" {
			rf = "(" + "~" + left + "|" + right + ")"
			return rf
		}
		rf = left + op + right
		if op == pop {
			return rf
		}
		return "(" + rf + ")"
	case NotOpExpr:
		op := string(rune(e.(NotOpExpr).Operator))
		right := convImply(e.(NotOpExpr).Right, op)
		return op + right
	case Atomic:
		return e.(Atomic).Atomic
	default:
		return ""
	}
}

// convNeg は構文木を再帰的に探索し, 否定をドモルガンの法則に従って変数へ寄せる. また,二重否定を削除する.
func convNeg(e /* expression */ Expression, pop /* parent op */ string, nc /* negation counter */ int) string {
	var rf string
	switch e.(type) {
	case BinOpExpr:
		op := string(rune(e.(BinOpExpr).Operator))
		left := convNeg(e.(BinOpExpr).Left, op, nc)
		right := convNeg(e.(BinOpExpr).Right, op, nc)
		// 二重否定の場合は, &と|の変換を行わない.
		if nc%2 != 0 {
			if op == "&" {
				rf = left + "|" + right
			}
			if op == "|" {
				rf = left + "&" + right
			}
		} else {
			rf = left + op + right
		}

		if op == pop {
			return rf
		}
		return "(" + rf + ")"
	case NotOpExpr:
		op := string(rune(e.(NotOpExpr).Operator))
		// 構文木の根方向にいくつ否定があるかを葉方向のノードが把握できるように,否定数をインクリメントする.
		nc++
		right := convNeg(e.(NotOpExpr).Right, op, nc)
		return right
	case Atomic:
		rf = e.(Atomic).Atomic
		// 二重否定の場合は否定記号を追加しない.
		if nc%2 != 0 {
			rf = "~" + rf
		}
		return rf
	default:
		return ""
	}
}

// ConvTseitin はNNFを受け取りTseitin変換を行う.
func ConvTseitin(f /* formula */ string) ([][]string, error) {
	r := strings.NewReader(f)
	// goyaccで構文木を作成する.
	p, err := Parse(r)
	if err != nil {
		return nil, err
	}

	// 構文木の探索でfresh variableの生成器を再帰関数内で共有するため, tseitinConverterをインスタンス化する.
	tc := newTseitinConverter()
	// 再帰関数はtseitinConverterのメソッドとして実装しているため, 呼び出す.
	fl, _ := tc.conv(p, "")
	return fl, nil

}
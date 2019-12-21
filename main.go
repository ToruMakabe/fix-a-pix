package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-sat"
	gosatcnf "github.com/mitchellh/go-sat/cnf"
	"gonum.org/v1/gonum/stat/combin"
)

const inputFormatMsg = "Please input -1 or 0-9. A non-negative number means the number of surrounding boxes, and -1 means not specified. Like this\n-1 2 3 -1 -1 0 -1 -1 -1 -1\n-1 -1 -1 -1 3 -1 2 -1 -1 6\n-1 -1 5 -1 5 3 -1 5 7 4\n-1 4 -1 5 -1 5 -1 6 -1 3\n-1 -1 4 -1 5 -1 6 -1 -1 3\n-1 -1 -1 2 -1 5 -1 -1 -1 -1\n4 -1 1 -1 -1 -1 1 1 -1 -1\n4 -1 1 -1 -1 -1 1 -1 4 -1\n-1 -1 -1 -1 6 -1 -1 -1 -1 4\n-1 4 4 -1 -1 -1 -1 4 -1 -1\n"

// fixは実質的な主処理である.
func fix() int {
	flag.Usage = flagUsage
	flag.Parse()

	// 引数の有無を検証する.
	args := flag.Args()
	if len(args) != 1 {
		flagUsage()
		return 1
	}

	// 以降を処理時間の計測対象とする.
	st := time.Now()

	// 入力ファイルをパースし, 形式を検証する.
	input, err := parseProblem(args[0])
	if err != nil {
		printError(err)
		return 1
	}

	// パースされた問題を表示する.
	fmt.Println("[Input problem]")
	for _, row := range input {
		for _, n := range row {
			fmt.Printf("%v ", n)
		}
		fmt.Println()
	}
	fmt.Println()

	// 入力データの行, 列サイズを取得する.
	columnCount := len(input[0])
	rowCount := len(input)

	// 端や角を考慮し, 塗ることができるセルの番号をまとめた表を作る.
	var printableCellsTable [][][]int
	for i := 1; i <= rowCount; i++ {
		printableCellsTable = append(printableCellsTable, nil)
		for j := 1; j <= columnCount; j++ {
			printableCellsTable[i-1] = append(printableCellsTable[i-1], nil)
			var s []int
			for k := -1; k <= 1; k++ {
				for l := -1; l <= 1; l++ {
					if i+k > 0 && j+l > 0 && i+k <= rowCount && j+l <= columnCount {
						s = append(s, (i-1+k)*(columnCount)+j+l)
					}
				}
			}
			printableCellsTable[i-1][j-1] = append(printableCellsTable[i-1][j-1], s...)
		}
	}

	// 入力データを元に, 塗ることができるセルの全ての組み合わせを取得する.
	var allCombi [][][]int
	for i := range input {
		for j := range input[i] {
			if input[i][j] != -1 {
				num := input[i][j]
				printableCells := printableCellsTable[i][j]
				if len(printableCells) <= num {
					num = len(printableCells)
				}
				c := combinations(printableCells, num)

				// 組み合わせが1つ(全て正か負)であれば, 塗るか塗らないかが確定する.
				if len(c) == 1 {
					var r [][]int
					for _, v := range c[0] {
						r = append(r, []int{v})
					}
					allCombi = append(allCombi, r)
				} else {
					allCombi = append(allCombi, c)
				}
			}
		}
	}

	// Tseitin変換を行う. 入力形式を限定できるため, 簡易な実装とする.
	fv := rowCount * columnCount
	var cnf [][]int
	for _, c := range allCombi {
		var fvs []int
		if len(c[0]) == 1 {
			for _, i := range c {
				cnf = append(cnf, i)
			}
			continue
		} else {
			for _, i := range c {
				fv++
				fvs = append(fvs, fv)
				var nc [][]int
				for _, j := range i {
					v := []int{-fv, j}
					nc = append(nc, v)
				}
				cnf = append(cnf, nc...)
			}
		}
		cnf = append(cnf, fvs)
	}

	// CNFの大きさ(節数)を表示する.
	fmt.Printf("Number of generated CNF clauses: %v\n", len(cnf))

	// 充足可否と付値を取得する.
	// SATソルバーにはMITライセンスで公開されている go-sat を利用する(https://github.com/mitchellh/go-sat)
	formula := gosatcnf.NewFormulaFromInts(cnf)
	s := sat.New()
	s.AddFormula(formula)
	r := s.Solve()

	fmt.Printf("SAT: %v\n", r)
	as := s.Assignments()

	// 結果は端末の表示幅に収まらない可能性があるため, ファイルに出力する.
	filename := "sol-" + filepath.Base(args[0])
	output, err := os.Create(filename)
	if err != nil {
		printError(err)
		return 1
	}
	defer output.Close()

	// 行, 列方向に走査し, 対応する変数の真偽をもとにセルを塗るか塗らないかを判断する. 見栄えを考慮してセル幅は2とする.
	for i := 1; i <= rowCount; i++ {
		for j := 1; j <= columnCount; j++ {
			if as[(i-1)*columnCount+j] {
				_, err := output.Write([]byte("  "))
				if err != nil {
					printError(err)
					return 1
				}
			} else {
				_, err := output.Write([]byte("##"))
				if err != nil {
					printError(err)
					return 1
				}
			}
		}
		_, err := output.Write([]byte("\n"))
		if err != nil {
			printError(err)
			return 1
		}
	}

	// 処理時間を表示する.
	et := time.Now()
	fmt.Println("Time: ", et.Sub(st))

	return 0
}

// parseProblemはfix-a-pixの問題ファイルを受け取り, 形式を検証する.
func parseProblem(fn /* filename */ string) ([][]int, error) {
	re := regexp.MustCompile("^-1|^[0-9]")
	var input [][]int

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		l := scanner.Text()
		c := strings.Split(l, " ")
		var s []int
		for _, n := range c {
			if !re.MatchString(n) {
				return nil, fmt.Errorf(inputFormatMsg)
			}
			if n == "-1" {
				s = append(s, -1)
			} else {
				i, _ := strconv.Atoi(n)
				s = append(s, i)
			}
		}
		input = append(input, s)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	columnCount := 0
	for _, row := range input {
		if len(row) == 0 {
			return nil, fmt.Errorf(inputFormatMsg)
		}
		if columnCount == 0 || columnCount == len(row) {
			columnCount = len(row)
		} else {
			return nil, fmt.Errorf(inputFormatMsg)
		}
	}
	return input, nil
}

// combinationsはスライス要素の組み合わせ(nCk)を作る. 組み合わせに含まれない変数は負数とする.
func combinations(s /* slice */ []int, k /* k-combination */ int) [][]int {
	var rs [][]int
	cs := combin.Combinations(len(s), k)
	for _, c := range cs {
		var r []int
		for _, v := range s {
			for _, n := range c {
				if v == s[n] {
					r = append(r, s[n])
					goto loopBottom
				}
			}
			r = append(r, -v)
		loopBottom:
		}
		rs = append(rs, r)
	}
	return rs
}

// flagUsageはコマンドラインオプション(フラグ)の使い方を出力する.
func flagUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %[1]s <problem-filename>\n", os.Args[0])
	flag.PrintDefaults()
}

// printErrorはエラーメッセージ出力を統一する.
func printError(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
}

// mainはエントリーポイントと終了コードを返却する役割のみとする.
func main() {
	os.Exit(fix())
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ToruMakabe/fix-a-pix/formula"
	"gonum.org/v1/gonum/stat/combin"
)

const inputFormatMsg = "Please input \".\" or 1-9. \".\" is empty as Fix-a-Pix cell."

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

	// 入力ファイルをパースし, 形式を検証する.
	input, err := parseProblem(args[0])
	if err != nil {
		printError(err)
		return 1
	}

	// 以降を処理時間の計測対象とする.
	st := time.Now()

	// パースされた問題を表示する.
	fmt.Println("[Input problem]")
	for _, row := range input {
		for _, n := range row {
			fmt.Printf("%v ", n)
		}
		fmt.Println()
	}
	fmt.Println()

	columnCount := len(input[0])
	rowCount := len(input)

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

	var allCombi [][]int
	for i := range input {
		for j := range input[i] {
			if input[i][j] != -1 {
				num := input[i][j]
				if num != 0 {
					printableCells := printableCellsTable[i][j]
					c := combinations(printableCells, num)
					allCombi = append(allCombi, c...)
				}

			}
		}
	}

	var allCombiA [][]string
	for _, c := range allCombi {
		var s []string
		for _, n := range c {
			s = append(s, strconv.Itoa(n))
		}
		allCombiA = append(allCombiA, s)
	}

	var dnf string
	for _, c := range allCombiA {
		dnf += "(" + strings.Join(c, "&") + ")|"
	}
	dnf = strings.TrimRight(dnf, "|")

	// 否定標準形(NNF)への変換を行う.
	nnf, err := formula.ConvNNF(dnf)
	if err != nil {
		fmt.Println()
		printError(err)
		fmt.Println()
		fmt.Println(inputFormatMsg)
		return 1
	}

	// Tseitin変換を行う.
	cnfm, err := formula.ConvTseitin(nnf)
	if err != nil {
		fmt.Println()
		printError(err)
		fmt.Println()
		fmt.Println(inputFormatMsg)
		return 1
	}
	fmt.Println(cnfm)

	// 処理時間を表示する.
	et := time.Now()
	fmt.Println("Time: ", et.Sub(st))

	return 0
}

// convCNFは入力された問題をCNFに変換する.
func convCNF(s /* input */ [][]string) ([][]string, error) {
	return s, nil
}

// parseProblemはfix-a-pixの問題ファイルを受け取り, 形式を検証する.
func parseProblem(fn /* filename */ string) ([][]int, error) {
	re := regexp.MustCompile("[0-9]|.")
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
			if n == "." {
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

// combinationsはスライス要素の組み合わせ(nCk)を作る.
func combinations(s /* slice */ []int, k /* k-combination */ int) [][]int {
	var rs [][]int
	cs := combin.Combinations(len(s), k)
	for _, c := range cs {
		var r []int
		for _, n := range c {
			r = append(r, s[n])
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

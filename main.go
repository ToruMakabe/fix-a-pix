package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

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

	rowLength := len(input[0])
	rowCounter := len(input)

	var paintableCells [][][]int

	for i := 1; i <= rowCounter; i++ {
		paintableCells = append(paintableCells, nil)
		for j := 1; j <= rowLength; j++ {
			paintableCells[i-1] = append(paintableCells[i-1], nil)
			s := []int{j + ((i - 1) * rowLength)}
			switch j {
			case 1:
				s = append(s, j+1+((i-1)*rowLength))
				switch i {
				case 1:
					s = append(s, j+1+(i*rowLength))
				case rowCounter:
					s = append(s, j+1+((i-2)*rowLength))
				default:
					s = append(s, j+1+(i*rowLength), j+1+((i-2)*rowLength))
				}
			case rowLength:
				s = append(s, j-1+((i-1)*rowLength))
				switch i {
				case 1:
					s = append(s, j-1+(i*rowLength))
				case rowCounter:
					s = append(s, j-1+((i-2)*rowLength))
				default:
					s = append(s, j-1+(i*rowLength), j-1+((i-2)*rowLength))
				}
			default:
				s = append(s, j-1+((i-1)*rowLength), j+1+((i-1)*rowLength))
				switch i {
				case 1:
					s = append(s, j-1+(i*rowLength), j+1+(i*rowLength))
				case rowCounter:
					s = append(s, j-1+((i-2)*rowLength), j+1+((i-2)*rowLength))
				default:
					s = append(s, j-1+(i*rowLength), j+1+(i*rowLength), j-1+((i-2)*rowLength), j+1+((i-2)*rowLength))
				}
			}
			switch i {
			case 1:
				s = append(s, j+((i-1)*rowLength)+rowLength)
			case rowCounter:
				s = append(s, j+((i-1)*rowLength)-rowLength)
			default:
				s = append(s, j+((i-1)*rowLength)+rowLength, j+((i-1)*rowLength)-rowLength)
			}
			paintableCells[i-1][j-1] = append(paintableCells[i-1][j-1], s...)
		}
	}

	fmt.Println(paintableCells)

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
func parseProblem(fn /* filename */ string) ([][]string, error) {
	re := regexp.MustCompile("[0-9]|.")
	var input [][]string

	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		l := scanner.Text()
		c := strings.Split(l, " ")
		var s []string
		for _, n := range c {
			if !re.MatchString(n) {
				return nil, fmt.Errorf(inputFormatMsg)
			}
			s = append(s, n)
		}
		input = append(input, s)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	rowLength := 0
	for _, row := range input {
		if len(row) == 0 {
			return nil, fmt.Errorf(inputFormatMsg)
		}
		if rowLength == 0 || rowLength == len(row) {
			rowLength = len(row)
		} else {
			return nil, fmt.Errorf(inputFormatMsg)
		}
	}

	return input, nil
}

// combinationsはスライス要素の組み合わせ(nC2)を作り, かつ否定の選言を表現するため各要素を負数に変換する.
func combinations(s /* slice */ []int) [][]int {
	var r [][]int
	cs := combin.Combinations(len(s), 2)
	for _, c := range cs {
		t := []int{-s[c[0]], -s[c[1]]}
		r = append(r, t)
	}
	return r
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

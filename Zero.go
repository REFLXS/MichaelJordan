package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)


func main() {
	rand.Seed(time.Now().UnixNano())

	// Создание матрицы (без исходных значений)
	matrix, headersRow, headersCol := CreateMatrix(nil)
	fmt.Println("Исходная матрица:")
	PrintMatrix(matrix, headersRow, headersCol)

	// Добавляем строку и столбец
	matrix, headersCol = AddRow(matrix, headersCol)
	matrix, headersRow = AddCell(matrix, headersRow)
	fmt.Println("\nПосле добавления строки и столбца:")
	PrintMatrix(matrix, headersRow, headersCol)

	fmt.Println("\nРешение методом Жордана:")
	matrix = JordanMethod(matrix, headersRow, headersCol)
}


func CreateMatrix(array [][]float64) ([][]float64, []string, []string) {
	sizeRow, sizeCol := 4, 5
	if array != nil {
		sizeRow = len(array) + 1
		sizeCol = len(array[0]) + 1
	}

	matrix := make([][]float64, sizeRow)
	headersRow := make([]string, sizeRow)
	headersCol := make([]string, sizeCol)

	for i := 0; i < sizeRow; i++ {
		matrix[i] = make([]float64, sizeCol)
		for j := 0; j < sizeCol; j++ {
			if i == 0 {
				if j == 1 {
					matrix[i][j] = 1
				} else if j != 0 {
					matrix[i][j] = math.NaN() // заголовок
				}
				if j > 0 {
					headersCol[j] = fmt.Sprintf("x%d", j)
				}
			} else {
				if j == 0 {
					matrix[i][j] = 0
				} else {
					if array == nil {
						matrix[i][j] = getRandom()
					} else {
						matrix[i][j] = array[i-1][j-1]
					}
				}
				headersRow[i] = fmt.Sprintf("y%d", i)
			}
		}
	}
	return matrix, headersRow, headersCol
}

func getRandom() float64 {
	return float64(rand.Intn(21) - 10)
}


func AddRow(matrix [][]float64, headersCol []string) ([][]float64, []string) {
	newRow := make([]float64, len(matrix[0]))
	newRow[0] = 0
	for i := 1; i < len(newRow); i++ {
		newRow[i] = getRandom()
	}
	matrix = append(matrix, newRow)
	return matrix, headersCol
}

func AddCell(matrix [][]float64, headersRow []string) ([][]float64, []string) {
	for i := range matrix {
		if i == 0 {
			matrix[i] = append(matrix[i], math.NaN())
		} else {
			matrix[i] = append(matrix[i], getRandom())
		}
	}
	headersRow = append(headersRow, fmt.Sprintf("x%d", len(headersRow)))
	return matrix, headersRow
}


func PrintMatrix(matrix [][]float64, headersRow []string, headersCol []string) {
	fmt.Printf("%-5s", "")
	for j := 1; j < len(headersCol); j++ {
		fmt.Printf("%10s", headersCol[j])
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 10*len(headersCol)))

	for i := 1; i < len(matrix); i++ {
		fmt.Printf("%-5s", headersRow[i])
		for j := 1; j < len(matrix[i]); j++ {
			fmt.Printf("%10.3f", matrix[i][j])
		}
		fmt.Println()
	}
}


func JordanMethod(arr [][]float64, headersRow, headersCol []string) [][]float64 {
	rank := MatrixRank(arr)
	arrAe := []int{} // индексы уже использованных разрешающих элементов

	for step := 1; step <= rank; step++ {
		iAe, jAe := ChoosePivot(arr, arrAe)
		fmt.Printf("\nЭтап %d: разрешающий элемент arr[%d][%d] = %.3f\n", step, iAe, jAe, arr[iAe][jAe])
		arr = JordanStep(arr, iAe, jAe)
		arrAe = append(arrAe, jAe)
		PrintMatrix(arr, headersRow, headersCol)
	}

	fmt.Println("\nИтоговая матрица после метода Жордана:")
	PrintMatrix(arr, headersRow, headersCol)
	return arr
}

func ChoosePivot(arr [][]float64, usedCols []int) (int, int) {
	for i := 1; i < len(arr); i++ {
		for j := 1; j < len(arr[i]); j++ {
			if arr[i][j] != 0 && !contains(usedCols, j) {
				return i, j
			}
		}
	}
	panic("Не найден разрешающий элемент")
}

func contains(arr []int, val int) bool {
	for _, v := range arr {
		if v == val {
			return true
		}
	}
	return false
}

func JordanStep(arr [][]float64, iAe, jAe int) [][]float64 {
	allowEntry := arr[iAe][jAe]
	if allowEntry == 0 {
		panic(fmt.Sprintf("Ошибка: разрешающий элемент arr[%d][%d] = 0", iAe, jAe))
	}

	m := len(arr)
	n := len(arr[0])
	newArr := make([][]float64, m)
	for i := 0; i < m; i++ {
		newArr[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			switch {
			case i == iAe && j == jAe:
				newArr[i][j] = 1 / allowEntry
			case i == iAe && j != jAe:
				newArr[i][j] = -arr[i][j] / allowEntry
			case j == jAe && i != iAe:
				newArr[i][j] = arr[i][j] / allowEntry
			default:
				newArr[i][j] = (arr[i][j]*allowEntry - arr[i][jAe]*arr[iAe][j]) / allowEntry
			}
			if math.IsNaN(newArr[i][j]) || math.IsInf(newArr[i][j], 0) {
				newArr[i][j] = 0
			}
		}
	}
	return newArr
}

func MatrixRank(A [][]float64) int {
	m := len(A)
	n := len(A[0])
	k := m
	if n < m {
		k = n
	}
	rank := 0
	for r := 1; r <= k; r++ {
		B := make([][]float64, r)
		for i := range B {
			B[i] = make([]float64, r)
		}
		for a := 0; a < m-r+1; a++ {
			for b := 0; b < n-r+1; b++ {
				for c := 0; c < r; c++ {
					for d := 0; d < r; d++ {
						B[c][d] = A[a+c][b+d]
					}
				}
				if Determinant(B) != 0 {
					rank = r
				}
			}
		}
	}
	return rank
}

func Determinant(A [][]float64) float64 {
	N := len(A)
	B := make([][]float64, N)
	for i := 0; i < N; i++ {
		B[i] = make([]float64, N)
		copy(B[i], A[i])
	}

	denom := 1.0
	exchanges := 0
	for i := 0; i < N-1; i++ {
		maxN := i
		maxValue := math.Abs(B[i][i])
		for j := i + 1; j < N; j++ {
			if math.Abs(B[j][i]) > maxValue {
				maxN = j
				maxValue = math.Abs(B[j][i])
			}
		}
		if maxN > i {
			B[i], B[maxN] = B[maxN], B[i]
			exchanges++
		} else if maxValue == 0 {
			return 0
		}
		value1 := B[i][i]
		for j := i + 1; j < N; j++ {
			value2 := B[j][i]
			B[j][i] = 0
			for k := i + 1; k < N; k++ {
				B[j][k] = (B[j][k]*value1 - B[i][k]*value2) / denom
			}
		}
		denom = value1
	}
	if exchanges%2 == 1 {
		return -B[N-1][N-1]
	}
	return B[N-1][N-1]
}

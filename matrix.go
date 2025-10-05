package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	resetState(3, 3)
}

func resetState(rows, cols int) {
	mat := make(Matrix, rows)
	for i := range mat {
		mat[i] = make([]float64, cols+1)
		for j := range mat[i] {
			mat[i][j] = float64(rand.Intn(21) - 10)
		}
	}
	current = mat
	rowLabels = make([]string, rows)
	for i := 0; i < rows; i++ {
		rowLabels[i] = "0"
	}
	colLabels = make([]string, cols+1)
	colLabels[0] = "1"
	for i := 1; i < cols+1; i++ {
		colLabels[i] = fmt.Sprintf("x%d", i)
	}
}

func cloneMatrix(a Matrix) Matrix {
	copyMat := make(Matrix, len(a))
	for i := range a {
		copyMat[i] = append([]float64{}, a[i]...)
	}
	return copyMat
}

func cloneSlice(s []string) []string {
	return append([]string{}, s...)
}

func MatrixRank(a Matrix) int {
	rows := len(a)
	if rows == 0 {
		return 0
	}
	cols := len(a[0])

	return int(math.Min(float64(rows), float64(cols-1)))
}

func allowEntry(arr []float64, arr_ae []int) int {
	for i := 1; i < len(arr); i++ {
		skipped := false
		for _, v := range arr_ae {
			if i == v {
				skipped = true
				break
			}
		}
		if skipped {
			continue
		}
		if math.Abs(arr[i]) > 1e-9 {
			return i
		}
	}
	return -1
}

func JordanovaException(arr Matrix, rowLabels, colLabels []string, i_ae, j_ae int) (Matrix, []string, []string) {
	allow_entry := arr[i_ae][j_ae]
	if math.Abs(allow_entry) < 1e-12 {
		return arr, rowLabels, colLabels
	}

	array := make(Matrix, len(arr))
	newRowLabels := cloneSlice(rowLabels)
	newColLabels := cloneSlice(colLabels)

	newRowLabels[i_ae], newColLabels[j_ae] = colLabels[j_ae], rowLabels[i_ae]

	for i := 0; i < len(arr); i++ {
		array[i] = make([]float64, len(arr[i]))
		for j := 0; j < len(arr[i]); j++ {
			if i == i_ae && j == j_ae {
				array[i][j] = 1.0 / allow_entry
			} else if i == i_ae && j != j_ae {
				array[i][j] = -arr[i][j] / allow_entry
			} else if j == j_ae && i != i_ae {
				array[i][j] = arr[i][j] / allow_entry
			} else {
				array[i][j] = (arr[i][j]*allow_entry - arr[i][j_ae]*arr[i_ae][j]) / allow_entry
			}

			array[i][j] = math.Round(array[i][j]*1000) / 1000
		}
	}

	return array, newRowLabels, newColLabels
}

func solveSteps(a Matrix, rLabels, cLabels []string) ([]Step, error) {
	steps := []Step{}
	workingMatrix := cloneMatrix(a)
	workingRowLabels := cloneSlice(rLabels)
	workingColLabels := cloneSlice(cLabels)

	rank := MatrixRank(workingMatrix)
	arr_ae := []int{}

	for i := 1; i <= rank; i++ {
		j := allowEntry(workingMatrix[i-1], arr_ae)
		if j == -1 {
			found := false
			for k := i; k < len(workingMatrix); k++ {
				j = allowEntry(workingMatrix[k], arr_ae)
				if j != -1 {

					workingMatrix[i-1], workingMatrix[k] = workingMatrix[k], workingMatrix[i-1]
					workingRowLabels[i-1], workingRowLabels[k] = workingRowLabels[k], workingRowLabels[i-1]
					found = true
					break
				}
			}
			if !found {
				break
			}
		}

		arr_ae = append(arr_ae, j)

		desc := fmt.Sprintf("Этап №%d. Разрешающий элемент: %.3f[%d][%d]",
			i, workingMatrix[i-1][j], i, j)

		nextMatrix, nextRowLabels, nextColLabels := JordanovaException(
			workingMatrix, workingRowLabels, workingColLabels, i-1, j,
		)

		steps = append(steps, Step{
			Desc:      desc,
			Matrix:    cloneMatrix(nextMatrix),
			RowLabels: cloneSlice(nextRowLabels),
			ColLabels: cloneSlice(nextColLabels),
			PivotRow:  i - 1,
			PivotCol:  j,
		})

		workingMatrix = nextMatrix
		workingRowLabels = nextRowLabels
		workingColLabels = nextColLabels
	}

	if len(steps) > 0 {
		finalMatrix, finalColLabels := removeZeroColumns(workingMatrix, workingColLabels)

		if len(finalMatrix[0]) != len(workingMatrix[0]) {
			steps = append(steps, Step{
				Desc:      "Итоговая матрица",
				Matrix:    finalMatrix,
				RowLabels: cloneSlice(workingRowLabels),
				ColLabels: finalColLabels,
				PivotRow:  -1,
				PivotCol:  -1,
			})
		}
	}

	return steps, nil
}

func getEguations(arr Matrix, rowLabels, colLabels []string, rank int) []string {
	result := []string{}

	for i := 1; i <= rank; i++ {

		if !strings.HasPrefix(rowLabels[i-1], "x") {
			continue
		}

		equation := fmt.Sprintf("%s = ", rowLabels[i-1])
		constant := 0.0
		terms := []string{}

		for j := 0; j < len(colLabels); j++ {
			coef := arr[i-1][j]
			lbl := colLabels[j]

			if lbl == "1" {
				constant += coef
				continue
			}

			if strings.HasPrefix(lbl, "x") {
				isBasic := false
				for _, rl := range rowLabels {
					if rl == lbl {
						isBasic = true
						break
					}
				}
				if isBasic {
					continue
				}
				if math.Abs(coef) > 1e-9 {
					sign := "+"
					if coef < 0 {
						sign = "-"
					}
					term := fmt.Sprintf(" %s %.3f%s", sign, math.Abs(coef), lbl)
					terms = append(terms, term)
				}
			}
		}

		if len(terms) == 0 {
			equation += fmt.Sprintf("%.3f", constant)
		} else {

			if math.Abs(constant) > 1e-9 {
				equation += fmt.Sprintf("%.3f", constant)
			} else {
				first := strings.TrimPrefix(terms[0], " + ")
				first = strings.TrimPrefix(first, " - ")
				if strings.HasPrefix(terms[0], " - ") {
					first = "-" + first
				}
				equation += first
				terms = terms[1:]
			}
			for _, t := range terms {
				equation += t
			}
		}

		result = append(result, equation)
	}

	return result
}

func getSolutionVector(steps []Step) []string {
	if len(steps) == 0 {
		return nil
	}
	finalStep := steps[len(steps)-1]
	rank := len(finalStep.Matrix)

	return getEguations(finalStep.Matrix, finalStep.RowLabels, finalStep.ColLabels, rank)
}

func removeZeroColumns(matrix Matrix, colLabels []string) (Matrix, []string) {
	if len(matrix) == 0 || len(colLabels) == 0 {
		return matrix, colLabels
	}

	rows := len(matrix)
	cols := len(colLabels)
	keep := []int{}

	for j := 0; j < cols; j++ {
		if colLabels[j] != "0" {
			keep = append(keep, j)
		}
	}

	newMat := make(Matrix, rows)
	for i := range matrix {
		newMat[i] = make([]float64, len(keep))
		for k, j := range keep {
			newMat[i][k] = matrix[i][j]
		}
	}

	newLabels := make([]string, len(keep))
	for k, j := range keep {
		newLabels[k] = colLabels[j]
	}

	return newMat, newLabels
}

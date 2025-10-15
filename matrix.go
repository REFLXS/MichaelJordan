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

	if err := checkFinalZeroEquations(workingMatrix, workingRowLabels); err != nil {
		return steps, err
	}

	finalMatrix, finalRowLabels, finalColLabels := removeZeroColumnsAndRows(workingMatrix, workingRowLabels, workingColLabels)

	if err := checkFinalZeroEquations(finalMatrix, finalRowLabels); err != nil {
		return steps, err
	}

	if len(finalMatrix) != len(workingMatrix) || len(finalMatrix[0]) != len(workingMatrix[0]) {
		steps = append(steps, Step{
			Desc:      "Итоговая матрица",
			Matrix:    finalMatrix,
			RowLabels: finalRowLabels,
			ColLabels: finalColLabels,
			PivotRow:  -1,
			PivotCol:  -1,
		})
	}

	return steps, nil
}

func checkFinalZeroEquations(matrix Matrix, rowLabels []string) error {
	for i, row := range matrix {
		if rowLabels[i] == "0" {
			allZero := true
			for j := 1; j < len(row); j++ {
				if math.Abs(row[j]) > 1e-9 {
					allZero = false
					break
				}
			}
			if allZero && math.Abs(row[0]) > 1e-9 {
				return fmt.Errorf("система несовместна: получено уравнение 0 = %.3f", row[0])
			}
		}
	}
	return nil
}

func getEguations(arr Matrix, rowLabels, colLabels []string, rank int) []string {
	result := []string{}

	basicVars := make(map[string]bool)
	for _, rl := range rowLabels {
		if strings.HasPrefix(rl, "x") {
			basicVars[rl] = true
		}
	}

	freeVars := []string{}
	for _, cl := range colLabels {
		if strings.HasPrefix(cl, "x") && !basicVars[cl] && cl != "1" {
			freeVars = append(freeVars, cl)
		}
	}

	for i := 0; i < rank; i++ {
		if !strings.HasPrefix(rowLabels[i], "x") {
			continue
		}

		equation := fmt.Sprintf("%s = ", rowLabels[i])
		constant := arr[i][0]
		terms := []string{}

		for j := 1; j < len(colLabels); j++ {
			if !strings.HasPrefix(colLabels[j], "x") {
				continue
			}

			if basicVars[colLabels[j]] {
				continue
			}

			coef := arr[i][j]
			if math.Abs(coef) > 1e-9 {
				paramName := strings.Replace(colLabels[j], "x", "t", 1)
				sign := "+"
				absCoef := math.Abs(coef)
				if coef < 0 {
					sign = "-"
				}

				term := fmt.Sprintf(" %s %.3f·%s", sign, absCoef, paramName)
				terms = append(terms, term)
			}
		}

		if len(terms) == 0 {
			equation += fmt.Sprintf("%.3f", constant)
		} else {
			if math.Abs(constant) > 1e-9 {
				equation += fmt.Sprintf("%.3f", constant)
			}
			for _, term := range terms {
				equation += term
			}
			if math.Abs(constant) < 1e-9 && len(equation) > 0 {
				if strings.HasPrefix(equation, rowLabels[i]+" =  +") {
					equation = strings.Replace(equation, rowLabels[i]+" =  +", rowLabels[i]+" = ", 1)
				} else if strings.HasPrefix(equation, rowLabels[i]+" =  -") {
					equation = strings.Replace(equation, rowLabels[i]+" =  -", rowLabels[i]+" = -", 1)
				}
			}
		}

		result = append(result, equation)
	}

	for _, fv := range freeVars {
		paramName := strings.Replace(fv, "x", "t", 1)
		result = append(result, fmt.Sprintf("%s = %s", fv, paramName))
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

func removeZeroColumnsAndRows(matrix Matrix, rowLabels, colLabels []string) (Matrix, []string, []string) {
	if len(matrix) == 0 || len(colLabels) == 0 {
		return matrix, rowLabels, colLabels
	}

	rows := len(matrix)
	cols := len(colLabels)

	keepCols := []int{}
	for j := 0; j < cols; j++ {
		if colLabels[j] != "0" {
			keepCols = append(keepCols, j)
		}
	}

	keepRows := []int{}
	for i := 0; i < rows; i++ {
		isZeroRow := true
		for _, j := range keepCols {
			if math.Abs(matrix[i][j]) > 1e-9 {
				isZeroRow = false
				break
			}
		}
		if !isZeroRow {
			keepRows = append(keepRows, i)
		}
	}

	newMat := make(Matrix, len(keepRows))
	for newI, oldI := range keepRows {
		newMat[newI] = make([]float64, len(keepCols))
		for newJ, oldJ := range keepCols {
			newMat[newI][newJ] = matrix[oldI][oldJ]
		}
	}

	newRowLabels := make([]string, len(keepRows))
	for k, i := range keepRows {
		newRowLabels[k] = rowLabels[i]
	}

	newColLabels := make([]string, len(keepCols))
	for k, j := range keepCols {
		newColLabels[k] = colLabels[j]
	}

	return newMat, newRowLabels, newColLabels
}

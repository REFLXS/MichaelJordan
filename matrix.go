package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Matrix [][]float64

type Step struct {
	Desc      string
	Matrix    Matrix
	RowLabels []string
	ColLabels []string
	PivotRow  int
	PivotCol  int
}

type PageData struct {
	Matrix         Matrix
	RowLabels      []string
	ColLabels      []string
	Steps          []Step
	Error          string
	SolutionVector []string
}

var (
	current   Matrix
	rowLabels []string
	colLabels []string
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

func jordanStep(inMatrix Matrix, inRowLabels, inColLabels []string, pivotRow, pivotCol int) (Matrix, []string, []string, error) {
	if pivotRow < 0 || pivotRow >= len(inMatrix) || pivotCol < 0 || pivotCol >= len(inMatrix[0]) {
		return nil, nil, nil, fmt.Errorf("неверные координаты разрешающего элемента")
	}

	pivot := inMatrix[pivotRow][pivotCol]
	if math.Abs(pivot) < 1e-9 {
		return nil, nil, nil, fmt.Errorf("разрешающий элемент в [%d, %d] равен нулю", pivotRow, pivotCol)
	}

	outMatrix := cloneMatrix(inMatrix)
	outRowLabels := cloneSlice(inRowLabels)
	outColLabels := cloneSlice(inColLabels)

	outRowLabels[pivotRow], outColLabels[pivotCol] = outColLabels[pivotCol], outRowLabels[pivotRow]
	tempMatrix := cloneMatrix(inMatrix)
	outMatrix[pivotRow][pivotCol] = 1.0 / pivot

	for j := 0; j < len(tempMatrix[0]); j++ {
		if j != pivotCol {
			outMatrix[pivotRow][j] = -tempMatrix[pivotRow][j] / pivot
		}
	}

	for i := 0; i < len(tempMatrix); i++ {
		if i != pivotRow {
			outMatrix[i][pivotCol] = -tempMatrix[i][pivotCol] / pivot
		}
	}

	for i := 0; i < len(tempMatrix); i++ {
		if i == pivotRow {
			continue
		}
		for j := 0; j < len(tempMatrix[0]); j++ {
			if j == pivotCol {
				continue
			}
			outMatrix[i][j] = tempMatrix[i][j] - (tempMatrix[i][pivotCol]*tempMatrix[pivotRow][j])/pivot
		}
	}

	return outMatrix, outRowLabels, outColLabels, nil
}

func solveSteps(a Matrix, rLabels, cLabels []string) ([]Step, error) {
	steps := []Step{}

	workingMatrix := cloneMatrix(a)
	workingRowLabels := cloneSlice(rLabels)
	workingColLabels := cloneSlice(cLabels)

	numIterations := int(math.Min(float64(len(workingMatrix)), float64(len(workingMatrix[0])-1)))

	for k := 0; k < numIterations; k++ {
		pivotRow, pivotCol := -1, -1

		if math.Abs(workingMatrix[k][k+1]) > 1e-9 {
			pivotRow, pivotCol = k, k+1
		} else {
			for i := k + 1; i < len(workingMatrix); i++ {
				if math.Abs(workingMatrix[i][k+1]) > 1e-9 {
					pivotRow, pivotCol = i, k+1
					break
				}
			}
			if pivotRow == -1 {
				for i := k; i < len(workingMatrix); i++ {
					for j := k + 1; j < len(workingMatrix[0]); j++ {
						if math.Abs(workingMatrix[i][j]) > 1e-9 {
							pivotRow, pivotCol = i, j
							break
						}
					}
					if pivotRow != -1 {
						break
					}
				}
			}
		}

		if pivotRow == -1 || pivotCol == -1 {
			steps = append(steps, Step{
				Desc:      fmt.Sprintf("Алгоритм остановлен на шаге %d: разрешающий элемент не найден", k+1),
				Matrix:    workingMatrix,
				RowLabels: workingRowLabels,
				ColLabels: workingColLabels,
			})
			return steps, nil
		}

		desc := fmt.Sprintf(
			"Шаг %d: Разрешающий элемент M[%d][%d] = %.2f. Меняем местами '%s' и '%s'.",
			k+1, pivotRow, pivotCol, workingMatrix[pivotRow][pivotCol],
			workingRowLabels[pivotRow], workingColLabels[pivotCol],
		)

		nextMatrix, nextRowLabels, nextColLabels, err := jordanStep(
			workingMatrix, workingRowLabels, workingColLabels, pivotRow, pivotCol,
		)
		if err != nil {
			steps = append(steps, Step{
				Desc:      fmt.Sprintf("Алгоритм остановлен на шаге %d: %v", k+1, err),
				Matrix:    workingMatrix,
				RowLabels: workingRowLabels,
				ColLabels: workingColLabels,
			})
			return steps, nil
		}

		steps = append(steps, Step{
			Desc:      desc,
			Matrix:    nextMatrix,
			RowLabels: nextRowLabels,
			ColLabels: nextColLabels,
			PivotRow:  pivotRow,
			PivotCol:  pivotCol,
		})

		workingMatrix = nextMatrix
		workingRowLabels = nextRowLabels
		workingColLabels = nextColLabels
	}

	return steps, nil
}

func getSolutionVector(steps []Step) []string {
	if len(steps) == 0 {
		return nil
	}
	finalStep := steps[len(steps)-1]

	solution := make(map[string]float64)

	for i, rowLabel := range finalStep.RowLabels {
		if strings.HasPrefix(rowLabel, "x") {
			solution[rowLabel] = finalStep.Matrix[i][0]
		}
	}

	result := []string{}
	for j := 1; j < len(finalStep.ColLabels); j++ {
		varName := fmt.Sprintf("x%d", j)
		if val, ok := solution[varName]; ok {
			result = append(result, fmt.Sprintf("%s = %.3f", varName, val))
		} else {
			result = append(result, fmt.Sprintf("%s = free", varName))
		}
	}
	return result
}

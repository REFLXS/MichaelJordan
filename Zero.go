package main

import (
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"strconv"
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
		pivotRow, pivotCol := k, k+1

		desc := fmt.Sprintf("Шаг %d: Разрешающий элемент M[%d][%d] = %.2f. Меняем местами '%s' и '%s'.",
			k+1, pivotRow, pivotCol, workingMatrix[pivotRow][pivotCol], workingRowLabels[pivotRow], workingColLabels[pivotCol])

		nextMatrix, nextRowLabels, nextColLabels, err := jordanStep(workingMatrix, workingRowLabels, workingColLabels, pivotRow, pivotCol)
		if err != nil {
			return nil, fmt.Errorf("ошибка на шаге %d: %w", k+1, err)
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
	solution := make([]string, 0)
	// Find which variable corresponds to which row
	for i, rowLabel := range finalStep.RowLabels {
		if strings.HasPrefix(rowLabel, "x") {
			value := finalStep.Matrix[i][0] // The value is in the first column (1)
			solution = append(solution, fmt.Sprintf("%s = %.3f", rowLabel, value))
		}
	}
	return solution
}

var tmpl = template.Must(template.New("page").Funcs(template.FuncMap{
	"formatFloat": func(f float64) string {
		return fmt.Sprintf("%.3f", f)
	},
	"formatHeader": func(s string) string {
		if s == "1" || s == "0" {
			return s
		}
		return "-" + s
	},
	"formatRowLabel": func(s string) string {
		if s == "0" {
			return "0"
		}
		return s
	},
}).Parse(`
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8">
	<title>Метод Жордана-Гаусса (полное исключение)</title>
	<style>
		body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; margin: 2em; background-color: #f4f7f9; color: #333; }
		h2 { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 10px; }
		.main-container { display: flex; align-items: flex-start; gap: 20px; }
		table { border-collapse: collapse; margin-top: 15px; margin-bottom: 20px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); background-color: #fff; }
		th, td { border: 1px solid #ddd; padding: 12px; text-align: center; min-width: 80px; }
		th { background-color: #ecf0f1; font-weight: 600; }
		td { background-color: #fff; }
		input { width: 70px; text-align: center; border: 1px solid #ccc; padding: 8px; border-radius: 4px; transition: border-color 0.3s; }
		input:focus { outline: none; border-color: #3498db; }
		.button-container { display: flex; flex-direction: column; gap: 10px; }
		.btn {
			padding: 10px 18px;
			font-size: 16px;
			cursor: pointer;
			border: none;
			border-radius: 5px;
			color: white;
			font-weight: bold;
			text-transform: uppercase;
			transition: background-color 0.3s, transform 0.1s;
			min-width: 150px;
		}
		.btn-solve { background-color: #2ecc71; }
		.btn-random { background-color: #3498db; }
		.btn-mod { background-color: #95a5a6; }
		.btn:hover { transform: translateY(-2px); box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
		.btn-solve:hover { background-color: #27ae60; }
		.btn-random:hover { background-color: #2980b9; }
		.btn-mod:hover { background-color: #7f8c8d; }
		.step { margin: 30px 0; border-top: 2px solid #bdc3c7; padding-top: 20px; }
		.pivot { background-color: #f1c40f; font-weight: bold; }
		.error { color: #e74c3c; font-weight: bold; margin-top: 15px; padding: 10px; background-color: #fdedec; border-left: 5px solid #e74c3c; }
		.solution-vector {
			margin-top: 20px;
			font-size: 1.2em;
			font-weight: bold;
			background-color: #ecf0f1;
			padding: 15px;
			border-radius: 8px;
			border: 1px solid #bdc3c7;
		}
	</style>
</head>
<body>
<h2>Исходная матрица</h2>

<form method="POST">
<div class="main-container">
	<div>
		<table>
		<thead>
		<tr>
			<th></th>
			{{range .ColLabels}}
				<th>{{formatHeader .}}</th>
			{{end}}
		</tr>
		</thead>
		<tbody>
		{{range $i, $row := .Matrix}}
		<tr>
			<th>{{formatRowLabel (index $.RowLabels $i)}}=</th>
			{{range $j, $val := $row}}
				<td><input name="cell_{{$i}}_{{$j}}" value="{{printf "%.2f" $val}}"></td>
			{{end}}
		</tr>
		{{end}}
		</tbody>
		</table>
	</div>

	<div class="button-container">
		<button class="btn btn-solve" type="submit" name="action" value="solve">Решить</button>
		<button class="btn btn-random" type="submit" name="action" value="random">Случайная</button>
		<button class="btn btn-mod" type="submit" name="action" value="addrow">+ строка</button>
		<button class="btn btn-mod" type="submit" name="action" value="delrow">- строка</button>
		<button class="btn btn-mod" type="submit" name="action" value="addcol">+ столбец</button>
		<button class="btn btn-mod" type="submit" name="action" value="delcol">- столбец</button>
	</div>
</div>
</form>

{{if .Error}}
	<div class="error">Ошибка: {{.Error}}</div>
{{end}}

{{if .Steps}}
<h2>Пошаговое решение:</h2>
{{range .Steps}}
	{{ $step := . }}
	<div class="step">
		<b>{{$step.Desc}}</b>
		<table>
		<thead>
		<tr>
			<th></th>
			{{range $step.ColLabels}}
				<th>{{formatHeader .}}</th>
			{{end}}
		</tr>
		</thead>
		<tbody>
		{{range $i, $row := $step.Matrix}}
		<tr>
			<th>{{formatRowLabel (index $step.RowLabels $i)}}=</th>
			{{range $j, $val := $row}}
				<td class="{{if and (eq $i $step.PivotRow) (eq $j $step.PivotCol)}}pivot{{end}}">
					{{formatFloat $val}}
				</td>
			{{end}}
		</tr>
		{{end}}
		</tbody>
		</table>
	</div>
{{end}}

{{if .SolutionVector}}
	<div class="solution-vector">
		<h3>Вектор-решение:</h3>
		<p>[{{range $i, $val := .SolutionVector}}{{if $i}}, {{end}}{{$val}}{{end}}]</p>
	</div>
{{end}}
{{end}}

</body>
</html>
`))

func handler(w http.ResponseWriter, r *http.Request) {
	pageData := PageData{
		Matrix:    current,
		RowLabels: rowLabels,
		ColLabels: colLabels,
	}

	if r.Method == http.MethodPost {
		r.ParseForm()

		action := r.FormValue("action")
		if action != "random" {
			for i := range current {
				for j := range current[i] {
					name := fmt.Sprintf("cell_%d_%d", i, j)
					if v := r.FormValue(name); v != "" {
						v = strings.Replace(v, ",", ".", -1)
						if num, err := strconv.ParseFloat(v, 64); err == nil {
							current[i][j] = num
						}
					}
				}
			}
		}

		switch action {
		case "random":
			resetState(len(current), len(current[0])-1)
		case "addrow":
			newRow := make([]float64, len(current[0]))
			current = append(current, newRow)
			rowLabels = append(rowLabels, "0")
		case "delrow":
			if len(current) > 1 {
				current = current[:len(current)-1]
				rowLabels = rowLabels[:len(rowLabels)-1]
			}
		case "addcol":
			for i := range current {
				current[i] = append(current[i], 0)
			}
			colLabels = append(colLabels, fmt.Sprintf("x%d", len(colLabels)))
		case "delcol":
			if len(current[0]) > 2 {
				for i := range current {
					current[i] = current[i][:len(current[i])-1]
				}
				colLabels = colLabels[:len(colLabels)-1]
			}
		case "solve":
			steps, err := solveSteps(current, rowLabels, colLabels)
			if err != nil {
				pageData.Error = err.Error()
			} else {
				pageData.Steps = steps
				pageData.SolutionVector = getSolutionVector(steps)
			}
		}

		pageData.Matrix = current
		pageData.RowLabels = rowLabels
		pageData.ColLabels = colLabels
	}

	tmpl.Execute(w, pageData)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Сервер запущен: http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

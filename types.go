package main

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
	FinalMatrix    Matrix
	FinalRowLabels []string
	FinalColLabels []string
}

var (
	current   Matrix
	rowLabels []string
	colLabels []string
)

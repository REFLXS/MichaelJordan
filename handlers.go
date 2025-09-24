package main

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var tmpl = template.Must(template.New("page.html").Funcs(template.FuncMap{
	"formatFloat": func(f float64) string {
		return strconv.FormatFloat(f, 'f', 3, 64)
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
}).ParseFiles("page.html"))

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
					name := "cell_" + strconv.Itoa(i) + "_" + strconv.Itoa(j)
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
			colLabels = append(colLabels, "x"+strconv.Itoa(len(colLabels)))
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

package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	/*"sort"*/
	/*"exec"*/
	"fmt"
)

var (
	templates = template.Must(template.ParseFiles("tmpl/edit.html","tmpl/view.html", "tmpl/index.html"))
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

type Page struct {
	Title string
	Body []byte
}

// GIANT RUBIFY CODE

func rubify(code string) string {
	// Keep code intact
	temp := code

	// Declare arrays (find better way?)
	// It seems that a lot of this is important due to the complexity of Go method
	// declarations and the importance Go places on methods vs. Objects
	funcArr := make([]string, 0)
	cmdArr := make([]string, 0)
	funcBodyArr := make([]string, 0)
	funcParamArr := make([]string, 0)
	//funcInputArr := make([]string, 0)
	funcReturnArr := make([]string, 0)

	// 

	// Inside of function
	if (strings.Contains(temp, "function")) {

		// Partition the Ruby file at instances of new classes
		temp_ := strings.Split(temp, "functions ")

		for _, function := range temp_[1:] {
			//break function up
			function = strings.TrimSpace(function)
			f := string(function[(len(function)-4):])

			//if (strings.Contains(function, "def")) {
			// "end" signals the end of many things, such as a function
			if strings.EqualFold(strings.TrimSpace(f), "end") == true {
				func1 := function
				func1Arr := uncomment(strings.Split(func1, "\n"))

				name, iType := getHeader(func1Arr[0])
				body, ret := getInside(strings.Replace(func1, "\n", "|", -1))
				//body, ret := getInside(strings.Join(func1Arr[1:], "|"))

				funcArr = append(funcArr, name)
				funcBodyArr = append(funcBodyArr, body)
				funcParamArr = append(funcParamArr, iType)
				//funcInputArr = append(funcInputArr, iType)
				funcReturnArr = append(funcReturnArr, ret)
			} else {
				//if strings.Contains()
			}
		}
	// No methods in file
	} else {
		func1 := strings.TrimSpace(temp)
		//func1Arr := uncomment(strings.Split(func1, "\n"))

		//body,_ := getInside(strings.Join(func1Arr[0:], "|"))
		body,_ := getInside(strings.Replace(func1, "\n", "|", -1))
		fmt.Println(body)

		cmdArr = append(cmdArr, body)
		funcArr = append(funcArr, "")
		funcBodyArr = append(funcBodyArr, "")
		funcParamArr = append(funcParamArr, "")
		funcReturnArr = append(funcReturnArr, "")
	}

	// PRINTING CODE

	printCode := `package main

import (
`
	// For a Rails app, print "fmt" regardless of puts
	if strings.Contains(temp, "puts") {
		printCode += `	"fmt"
`
	}
	if strings.Contains(temp, "require File.") {
		printCode += `	"io/ioutil"
`
	}
	if strings.Contains(temp, "Sort") || strings.Contains(temp, "Reverse") {
		printCode += `	"sort"
`
	}
	// In case of rails and/or http requests
	if strings.Contains(temp, "html") {
		printCode += `	"html/template"
		"net/http"
		"regexp"
`
	}

	printCode += `)
`
	for index, function := range funcArr {
		if function != "" {
		if funcParamArr[index] == "" {
			printCode += `
func ` + function + `() ` + funcReturnArr[index] + `{
` + funcBodyArr[index] + `}
`
		} else {
			printCode += `
func ` + function + `(` + `varName ` + funcParamArr[index] + `) ` + funcReturnArr[index] + ` {
` + funcBodyArr[index] + `}
`
		}
		}
	}
	main := `
func main() {
`
	//main := ``
	for index, function := range funcArr {
		if function != "" {
			if funcParamArr[index] == "" {
				main += `	` + function + `()
`
			} else {
				main += `	` + function + `(` + "var" + `)
`
			}
		}
		if function=="" && cmdArr[index]!= "" {
			main += cmdArr[index] + `
`
		}
	}
	main += "}"

	printCode += main
	return printCode
}

// END OF RUBIFY

// FUNCTIONS RELATED TO RUBIFY

// NOTE: this method has different arrays
// Inputs a string separated by | where new lines are
// Returns a string of the inside of the function in Go
// Also returns a return type which may be useful for the function header
func getInside(input string) (string, string) {
	s := strings.TrimSpace(input)
	codeBody := ""
	space := `
`
	outType := ""
	if strings.Contains(s, "return") {
		outType = ""
	}
	// each string of the array is separated by |
	// s_arr contains each line of array in that order
	s_arr := strings.Split(s, "|")

	// initialize variables, their strings, and/or functions involved
	var_arr := make([]string, 0)
	string_arr := make([]string, 0)
	function_arr := make([]string, 0)

	// for each line (index) in array s_arr
	for num, index := range s_arr {
		index = strings.TrimSpace(index)
		//fmt.Println(index)
		if strings.Contains(index, `"`) && !strings.Contains(index, "puts") {
			//get string around quotes
			//str := strings.Split(index, `"`)
			string_arr = append(string_arr, "string")//str[1])
		} else {
			string_arr = append(string_arr, "")
		}
		if strings.Contains(index, "=") {
			function_arr = append(var_arr, strings.Replace(index, "=", ":=", -1))
		} else {
			var_arr = append(var_arr, "")
		}
		// puts -- similar to fmt.Println
		if strings.Contains(index, "puts") {
			puts_arr := strings.Split(index, " ")
			function_arr = append(function_arr, "fmt.Println(" + puts_arr[1] + ")")
		} else {
			// do nothing
			function_arr = append(function_arr, "")
		}
		// print -- similar to fmt.Printf
		if strings.Contains(index, "print") {
			puts_arr := strings.Split(index, " ")
			function_arr = append(function_arr, "fmt.Printf(" + puts_arr[1] + ")")
		} else {
			// do nothing
			function_arr = append(function_arr, "")
		}
		if (strings.Contains(index, "+")) || (strings.Contains(index, "-")) || (strings.Contains(index, "/")) || (strings.Contains(index, "*")) {
			function_arr = append(function_arr, "fmt.Println(" + index + ")")
		} else {
			function_arr = append(function_arr,"")
		}
		// Presence of conditionals
		// Conditionals are in the following format:
		// } else if /*conditional*/ {	and 	} else {
		if strings.Contains(index, "if") {
			function_arr = append(function_arr, index + ` {
`)
		}
		if strings.Contains(index, "else if") {
			function_arr = append(function_arr, index + ` {
`)
		}
		if strings.Contains(index, "else") {
			function_arr = append(function_arr, index + ` {`)
		}
		// Begin replacing loops
		// Go only has for loops
		if strings.Contains(index, "for") {
			// in the form: while $i < $j do
			// replace do with {
			function_arr = append(function_arr, index)
		}
		if strings.Contains(index, "while"){
			function_arr = append(function_arr, index)
		}
		// Check if this is end if conditionals or loops but not end of function
		// Note: for conditionals and loops
		if (index == "end" && num != len(s_arr)-1) {
			function_arr = append(function_arr, `
}`)
		}

	}

	// prints the inside of the function line by line
	for _, index := range function_arr {
		if index != "" {
			codeBody += "	" + index + space
		}
	}
	return codeBody, outType
}

func getHeader(input string) (string, string) {
	s := strings.TrimSpace(input)
	retVal := ""
	fName := ""
	if strings.Contains(s, " ") {
		tmp_arr := strings.Split(s, " ")
		fName = tmp_arr[0]
		retVal = "foo"
	} else if (strings.Contains(s, "(")) {
		retVal = "foo"
	} else {
		fName = s
	}
	//s_arr := strings.Split(s, " ")
	return fName, retVal
}

// Removes comments
func uncomment(input []string) []string {
	output := make([]string, 0)
	for _, ele := range input {
		ele = strings.TrimSpace(ele)
		// Replaces hashtag with double backslash
		if strings.Contains(ele, "#") {
			output = append(output, strings.Replace(ele, "#", "//",-1))
		} else {
			output = append(output, ele)
		}
		// Replace multiple-line comments
		// Complete removes lines that begin with hashtag (#)
		//if strings.EqualFold(string(ele)[0:1], "#") == false {
		//	output = append(output, ele)
		//}
	}
	return output
}

// END OF RELATED FUNCTIONS

// MISCELANEOUS RUBY FUNCTIONS
// Common functions include: reverse *, length  = len(), to_s = string(array[:]), to_i, to_a = strings.Split(string, ""),  max =
// *Functions listed here
// These functions should be printed out?

// Sort functions
// sort normall arr.sort
// Sort backwards in ruby: arr.sort {|x,y| y <=> x }
// sort.Reverse does completely diff things
func goSort(in_arr []string, s string) []string {
	sorted_arr := in_arr;
	if s == "{|x,y| y <=> x }" {
		//sorted_arr := sort.Sort(in_arr)
	} else {
		//sorted_arr := sort.Reverse(in_arr)
	}
	return sorted_arr
}

// Reverse function
// Note: This function returns the reverse of a string.
func stringReverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes) - 1; i < j; i, j = i + 1, j - 1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

// END OF RUBY FUNCTIONS

// BEGIN OVERHEAD GO CODE

func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func (p *Page) translate() error {
	ioutil.WriteFile("data/" + p.Title + ".rb", p.Body, 0600)
	filename := "data/" + p.Title + ".go"
	return ioutil.WriteFile(filename, ruby(p.Body), 0600)
}

func ruby(code []byte) []byte {
	c := rubify(string(code[:]))
	b := []byte(c)
	return b
}

func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".go"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	//err := p.save()
	err := p.translate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}


// Initial form entry
func enterHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("name")
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	title := r.URL.Path[len("/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "index", p)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/enter/", enterHandler)
	http.HandleFunc("/save/", makeHandler(saveHandler))
	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

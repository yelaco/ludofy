package main

import (
	"os"
	"text/template"
)

func main() {
	// Open the external template file
	tmplContent, err := os.ReadFile("./scripts/template/sam.tmpl")
	if err != nil {
		panic(err)
	}

	// Parse the template
	t := template.Must(template.New("sam").Parse(string(tmplContent)))

	// Template data — you can control this however you like
	data := map[string]bool{
		"IncludeStackA": true,
		"IncludeStackB": false,
	}

	// Create the output file
	out, err := os.Create("output/template.yaml")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Execute the template with data
	err = t.Execute(out, data)
	if err != nil {
		panic(err)
	}

	println("✅ template.yaml generated in /output")
}

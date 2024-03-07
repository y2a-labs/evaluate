package commands

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func executeTemplates(structName string, files []Data) {
	funcMap := template.FuncMap{
		"capitalize": capitalizeFirstLetter, // Assuming this function is defined elsewhere
		"lower":      lowerFirstLetter,      // Assuming this function is defined elsewhere
	}

	for _, file := range files {

		// Ensure the /generated directory exists
		if err := os.MkdirAll(file.Package, os.ModePerm); err != nil {
			panic(fmt.Sprintf("Failed to create directory %s: %v", file.Package, err))
		}
		tmpl, err := template.New(filepath.Base(file.TemplatePath)).Funcs(funcMap).ParseFiles(file.TemplatePath)
		if err != nil {
			panic(err) // Handle error appropriately in production code
		}

		filename := strings.ToLower(structName) + ".go"
		outputFilePath := filepath.Join(file.Package, filename)

		// Check if the output file already exists

		if _, err := os.Stat(outputFilePath); !os.IsNotExist(err) {
			fmt.Printf("File %s already exists, skipping...\n", outputFilePath)
			continue
		}

		// Create the output file
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			panic(fmt.Sprintf("Failed to create file %s: %v", outputFilePath, err))
		}
		defer outputFile.Close()

		// Execute the template and write the output to the file
		err = tmpl.Execute(outputFile, map[string]string{
			"StructName":  structName,
			"PackageName": file.Package,
		})
		if err != nil {
			panic(err) // Handle error appropriately in production code
		}
		fmt.Printf("Generated file: %s\n", outputFilePath)
	}
}

// capitalizeFirstLetter capitalizes the first letter of the string.
func capitalizeFirstLetter(str string) string {
	if str == "" {
		return ""
	}
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// capitalizeFirstLetter capitalizes the first letter of the string.
func lowerFirstLetter(str string) string {
	if str == "" {
		return ""
	}
	r := []rune(str)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

type Data struct {
	TemplatePath string
	OutputDir    string
	Package      string
}

func TraverseFolder(directory string) {
	serviceStructs := []string{"PromptHistory"}
	files := []Data{
		{
			TemplatePath: "./internal/templates/controller.go.tmpl",
			Package:      "api",
		},

		{
			TemplatePath: "./internal/templates/controller.go.tmpl",
			Package:      "web",
		},

		{
			TemplatePath: "./internal/templates/service.go.tmpl",
			Package:      "services",
		},
	}

	for _, structName := range serviceStructs {
		executeTemplates(structName, files)
	}
	/*
		err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error accessing path %q: %v\n", path, err)
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".go" {
				fmt.Printf("Processing file: %s\n", path)
				filename := filepath.Base(path)
				structName := strings.TrimSuffix(filename, ".go")
				executeTemplates(structName, files)
			}
			return nil
		})


		if err != nil {
			fmt.Printf("Error walking the path %q: %v\n", directory, err)
			return
		}
	*/
}

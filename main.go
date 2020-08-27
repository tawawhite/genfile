package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/BurntSushi/toml"
)

func main() {
	var varfile = flag.String("v", "", "the file contains variables, support file with json/toml extension")
	var input = flag.String("i", "", "the template file or the directory containing template files")
	var output = flag.String("o", "", "the output file or the directory to output files")
	var help = flag.Bool("h", false, "show help")
	flag.Parse()

	if *help || *varfile == "" || *input == "" || *output == "" {
		showHelp()
		return
	}

	vars, err := loadVarfile(*varfile)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	tmpls, err := loadTemplates(*input)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	for _, t := range tmpls.Templates() {
		fmt.Println("generating", t.Name(), "to", *output)
		if err := outputTemplate(*output, t, vars); err != nil {
			fmt.Println(err)
			os.Exit(3)
		}
	}
}

func showHelp() {
	fmt.Println(`Generate files from templates with given variables.

Usage: genfile -i <input> -v <variable file> -o <output>

Flags:`)
	flag.PrintDefaults()
	fmt.Println(`
Exit code:
  0	succeed
  1	help
  2	input error
  3	output error`)
	os.Exit(1)
}

func loadVarfile(varfile string) (interface{}, error) {
	f, err := os.Open(varfile)
	if err != nil {
		return nil, fmt.Errorf("load varfile error: %w", err)
	}
	defer f.Close()

	var ret interface{}
	ext := filepath.Ext(varfile)
	switch ext {
	case ".json":
		if err := json.NewDecoder(f).Decode(&ret); err != nil {
			return nil, fmt.Errorf("parse varfile as json error: %w", err)
		}
	case ".toml":
		if _, err := toml.DecodeReader(f, &ret); err != nil {
			return nil, fmt.Errorf("parse varfile as toml error: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid varfile. %s should be json/toml/yaml format, but %s.", varfile, ext)
	}

	return ret, nil
}

func loadTemplates(input string) (*template.Template, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, fmt.Errorf("load templates from %s error: can't determie if it's a directory: %w", input, err)
	}

	if !info.IsDir() {
		ret, err := template.ParseFiles(input)
		if err != nil {
			return nil, fmt.Errorf("load templates from %s error: %w", input, err)
		}
		return ret, nil
	}

	input = filepath.Clean(input)
	var ret *template.Template

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("path %s: %w", path, err)
		}
		if info.IsDir() {
			return nil
		}

		name := path[len(input)+1:]
		var t *template.Template
		if ret == nil {
			ret = template.New(name)
			t = ret
		} else {
			t = ret.New(name)
		}

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("load template %s error: %w", path, err)
		}
		if _, err := t.Parse(string(buf)); err != nil {
			return fmt.Errorf("parse template %s error: %w", path, err)
		}
		return nil
	}

	if err := filepath.Walk(input, walkFunc); err != nil {
		return nil, fmt.Errorf("load templates from %s error: %w", input, err)
	}

	return ret, nil
}

func outputTemplate(output string, t *template.Template, vars interface{}) error {
	file := filepath.Join(output, t.Name())

	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("create directory %s error: %w", dir, err)
	}

	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("create %s error: %w", file, err)
	}
	defer f.Close()

	if err := t.Execute(f, vars); err != nil {
		return fmt.Errorf("generate %s error: %w", file, err)
	}

	return nil
}

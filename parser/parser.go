package parser

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

//Intent is Parent struct for input intent
type Intent struct {
	DistIntent bool   `yaml:"distIntent"`
	Type       string `yaml:"type"`
	Location   []Location
	Sources    map[string]Specs
	Models     map[string]ModelSpecs
	Sinks      map[string]Specs
}

//Specs gives specifications for every element of intent
type Specs struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//ModelSpecs ...
type ModelSpecs struct {
	ID string `yaml:"id"`
	Req Requirements
}

//Requirements gives detailed reqs
type Requirements struct {
	Access       string `yaml:"access"`
	Size         string `yaml:"size"`
	Distribution string `yaml:"distribution"`
	Kind         string `yaml:"kind"`
}

//Location ... is
type Location struct {
	Server string `yaml:"server"`
}

//Parse parses the input yaml file
func Parse(filepath string) Intent {
	yamlfile, _ := ioutil.ReadFile(filepath)
	y := Intent{}
	err := yaml.Unmarshal([]byte(yamlfile), &y)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	//fmt.Printf("%+v\n", y)
	return y
}

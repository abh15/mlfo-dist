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
	Servers    []Server
	Sources    []Source
	Models     []Model
	Sinks      []Sink
}

//Source gives specifications for every element of intent
type Source struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//Sink ...
type Sink struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//Model ...
type Model struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//Requirements gives detailed reqs
type Requirements struct {
	Accuracy     string `yaml:"accuracy"`
	Size         string `yaml:"size"`
	Distribution string `yaml:"distribution"`
	Kind         string `yaml:"kind"`
	Num          int32    `yaml:"num"`
}

//Server ... is
type Server struct {
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

package parser

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Intent struct {
	DistIntent bool
	Sources    map[string]Specs
	Models     map[string]Specs
	Sinks      map[string]Specs
}

type Specs struct {
	Id           string
	Requirements map[string]string
	Constraints  map[string]string
}

//Parse parses the input yaml file
func Parse(filepath string) Intent {
	yamlfile, _ := ioutil.ReadFile(filepath)
	y := Intent{}
	err := yaml.Unmarshal([]byte(yamlfile), &y)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return y
}

package parser

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

//Intent is Parent struct for input intent
type Intent struct {
	DistIntent bool
	Sources    map[string]Specs
	Models     map[string]Specs
	Sinks      map[string]Specs
}

//Specs gives specifications for every element of intent
type Specs struct {
	ID  string `yaml:"id"`
	Req struct {
		Access       string `yaml:"access"`
		Server       string `yaml:"server"`
		Size         string `yaml:"size"`
		Federated    bool   `yaml:"federated"`
		Distribution string `yaml:"distribution"`
		Kind         string `yaml:"kind"`
	}
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

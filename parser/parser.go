package parser

import (
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

//Source specifies source requirements and/or ID
type Source struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//Sink specifies source requirements and/or ID
type Sink struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//Model specifies source requirements and/or ID
type Model struct {
	ID  string `yaml:"id"`
	Req Requirements
}

//Requirements gives detailed reqs for pipeline nodes
type Requirements struct {
	Accuracy     string `yaml:"accuracy"`
	Size         string `yaml:"size"`
	Distribution string `yaml:"distribution"`
	Kind         string `yaml:"kind"`
	Num          int32  `yaml:"num"`
}

//Server specifies server/CloudMLFO IP
type Server struct {
	Server string `yaml:"server"`
}

//Parse parses the input yaml file into the Intent{} struct
func Parse(yamlfile []byte) Intent {
	//	yamlfile, _ := ioutil.ReadFile(filepath)
	y := Intent{}
	err := yaml.Unmarshal([]byte(yamlfile), &y)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	//fmt.Printf("%+v\n", y)
	return y
}

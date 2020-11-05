package parser

import (
	"io/ioutil"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

//Parse parses the input yaml file
func Parse(filepath string) (interface{}, interface{}, interface{}, bool) {

	var sources interface{}
	var model interface{}
	var sinks interface{}
	var isDist bool

	yamlMap := make(map[interface{}]interface{})
	ymlContent, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(ymlContent, &yamlMap)
	for k, v := range yamlMap {
		switch k {
		case "distIntent":
			mirror := reflect.ValueOf(v)
			isDist = mirror.Interface().(bool)
		case "sources":
			sources = v
		case "models":
			model = v
		case "sinks":
			sinks = v
		}
	}

	return sources, model, sinks, isDist
}

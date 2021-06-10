package parser

import (
	"log"

	yaml "gopkg.in/yaml.v2"
)

//Intent is Parent struct for input intent
type Intent struct {
	IntentID string `yaml:"intentID"`
	Targets  []Target
	//FedServerIP string `yaml:"fedserverip"` //This is specific to the code and not part of general intent structure
}

//Target describes the desired actions for the target
type Target struct {
	ID          string `yaml:"id"`
	Operation   string `yaml:"operation"`
	Operand     string `yaml:"operand"`
	Constraints Constraints
}

//Constraints describes constraints for the action
type Constraints struct {
	Privacylevel string `yaml:"privacylevel"`
	Latency      string `yaml:"latency"`
	Sourcekind   string `yaml:"sourcekind"`
	Modelkind    string `yaml:"modelkind"`
	Minaccuracy  int32  `yaml:"minaccuracy"`
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

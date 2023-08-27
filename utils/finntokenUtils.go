package utils

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"reflect"
	"time"
)

//collect1: cfd5n71r01qj357er8v0cfd5n71r01qj357er8vg
//collect2: cfd5ne9r01qj357er950cfd5ne9r01qj357er95g
//collect3: cfd5nlhr01qj357er98gcfd5nlhr01qj357er990
//update: '20230201'

type TokenConfFormat struct {
	Collect1 string `yaml:"collect1"`
	Collect2 string `yaml:"collect2"`
	Collect3 string `yaml:"collect3"`
	Update   string
}

type FinnToken struct {
	token *TokenConfFormat
}

func (self *FinnToken) finnTokenConfigParser(yamlPath string) {
	self.token = new(TokenConfFormat)

	yamlfile, err := ioutil.ReadFile(yamlPath)

	if err != nil {
		log.Fatalln("fail to open token file")
	}

	err = yaml.Unmarshal(yamlfile, self.token)

	if err != nil {
		log.Fatalln("fail to unmarshal yaml")
	}
}

func (self *FinnToken) GetToken(yamlPath string) string {

	self.finnTokenConfigParser(yamlPath)
	rval := reflect.ValueOf(self.token).Elem()

	size := rval.NumField() - 1
	rand.Seed(time.Now().UnixNano())
	randint := rand.Intn(size)

	return rval.Field(randint).String()

}

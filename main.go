package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func main() {
	data, _ := ioutil.ReadFile("models.yml")

	m := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(m)

}

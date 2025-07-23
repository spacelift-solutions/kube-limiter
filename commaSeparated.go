package main

import "os"

type commaSeparated struct {
	str string
}

func newCommaSeparated() *commaSeparated {
	return &commaSeparated{str: ""}
}

func (c *commaSeparated) add(value string) {
	if c.str == "" {
		c.str = value
	} else {
		c.str += "," + value
	}
}

func (c *commaSeparated) writeToFile(fileName string) {
	f, err := os.Create(fileName)
	if err != nil {
		logger.Fatalf("Error creating file %s: %v", fileName, err)
	}
	defer f.Close()

	_, err = f.WriteString(c.str)
	if err != nil {
		logger.Fatalf("Error writing to file %s: %v", fileName, err)
	}
	logger.Debugf("Successfully wrote to file %s", fileName)
}

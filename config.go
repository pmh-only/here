package main

import "os"

func (*HereServer) getDataPath() string {
	dataPath, ok := os.LookupEnv("DATA_PATH")
	if !ok {
		dataPath = "/data"
	}

	return dataPath
}

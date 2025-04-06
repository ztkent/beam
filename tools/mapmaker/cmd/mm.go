package main

import (
	"github.com/ztkent/beam/tools/mapmaker/mapmaker"
)

func main() {
	mapMaker := mapmaker.NewMapMaker(1024, 768)
	mapMaker.Init()
	defer mapMaker.Close()

	// Reopen the last opened file if it exists
	if lastFile, err := mapmaker.LoadConfig(); err == nil && lastFile != "" {
		if err := mapMaker.LoadMap(lastFile); err != nil {
			println("Error loading last map:", err.Error())
		}
	}

	mapMaker.Run()
}

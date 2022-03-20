package main

import "os"

func main() {
	cs := ChartographerService{}
	cs.Initialize(os.Args[1], "chartas.db")
	cs.Run(":8080")
	_ = cs.DB.Close()
}

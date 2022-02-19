package main

func main() {
	cs := ChartographerService{}
	cs.Initialize()
	cs.Run(":8080")
}

package main

func main() {
	cs := ChartographerService{}
	cs.Initialize("chartas.db")
	cs.Run(":8080")
	_ = cs.DB.Close()
}

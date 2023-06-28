package main

func main() {
	ns := NewServer("127.0.0.1", 3333)
	ns.Start()
}

run: main.exe
	main.exe
main.exe:
	go build -o main.exe demo/ticket/main.go

run: main.exe
	main.exe
main.exe:
	go build -o main.exe demo/alipayFu/main.go
clear:
	rm main.exe

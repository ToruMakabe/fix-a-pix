APP=fix-a-pix

.PHONY:	 build
build: clean
	go build -o ${APP} main.go

.PHONY: clean
clean:
	go clean
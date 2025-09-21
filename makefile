build:
	go build -o chirpy .

run: build
	./chirpy

clean:
	rm -f chirpy

dev:
	go build -o chirpy . && ./chirpy

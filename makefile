build:
	go build -o chirpy .

run: build
	./chirpy

clean:
	rm -f chripy

dev:
	go build -o chripy . && ./chripy

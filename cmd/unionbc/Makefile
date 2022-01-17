FILES = main.go settings.go 

.PHONY: default build clean
default: clean build 
build: $(FILES)
	go build $(FILES)
clean:
	rm -rf main chain\:*

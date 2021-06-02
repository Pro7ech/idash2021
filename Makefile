# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get


# Binary names
BINARY_NAME=idash2021
BINARY_UNIX=$(BINARY_NAME)_unix


build:
	${GOBUILD} KeyGen.go
	${GOBUILD} ClientEnc.go
	${GOBUILD} ServerPred.go

key:
	./KeyGen 
enc:
	./ClientEnc 
pred:
	./ServerPred 

clean:
	${GOBUILD} Clean.go
	./Clean
	$(GOCLEAN)
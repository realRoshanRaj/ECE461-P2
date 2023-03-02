EXEC := executable
URL_TEST := https://github.com/nodejs/node
URL_FILE := URL_FILE.txt

help:
	@printf "%-20s %s\n" "------ Makefile Commands --------"
	@printf "%-20s %s\n" "Target" "Description"
	@printf "%-20s %s\n" "------" "-----------"
	@make -pqR : 2>/dev/null \
	| awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' \
	| sort \
	| egrep -v -e '^[^[:alnum:]]' -e '^$@$$' \
	| xargs -I _ sh -c 'printf "%-20s " _; make _ -nB | (grep -i "^# Help:" || echo "") | tail -1 | sed "s/^# Help: //g"'

install:
	@# Help: Runs setup commands
	go mod download

build:
	@# Help: Builds project
	go build -o ${EXEC} src/main.go

url: build
	./${EXEC} ${URL_FILE}

test: build
	@# Help: Runs test suite
	(cd src; go test -v -cover)
#go test -cover test_main.go

.PHONY: help install build git pull clean

EXEC := main
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
	@printf "%-20s %s\n" "------ ./run Commands -----------"
	./run help
	@printf "%-20s %s\n" "---------------------------------"

install:
	@# Help: Runs setup commands
	go get "github.com/joho/godotenv"
	go get "github.com/machinebox/graphql"
	go get "github.com/scizorman/go-ndjson"

build:
	@# Help: Builds project 
	go build -o ${EXEC} src/main.go

url: build
	./${EXEC} ${URL_FILE}

test:
	@# Help: Runs test suite 
	#./run test 
	

git: clean
	@# Help: Automates the git push workflow
	$(eval MESSAGE := $(shell bash -c 'read -p "commit -m " message; echo $$message'))
	@git add .
	@git status
	$(if $(strip $(MESSAGE)), git commit -m "$(MESSAGE) -$(shell whoami)", git commit -m "$(shell date +'%m-%d-%Y %H:%M:%S') -$(shell whoami)")
	@git pull
	@git push

pull:
	@# Help: Pulls from github
	@git pull

# EXAMPLE @rm -rf ________ || true
clean:
	@# Help: Removes unnecessary files 
	rm ${EXEC}

.PHONY: help install build git pull clean

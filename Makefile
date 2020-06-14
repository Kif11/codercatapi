dev:
	modd

build:
	env GOOS=linux GOARCH=amd64 go build github.com/kif11/codercatapi

deploy:
	make build
	ssh -t kiko@codercat.tk "sudo systemctl stop codercatapi"
	scp codercatapi kiko@codercat.tk:/home/kiko/codercatapi/
	ssh -t kiko@codercat.tk "sudo systemctl restart codercatapi"
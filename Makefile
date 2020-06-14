dev:
	modd

build:
	env GOOS=linux GOARCH=amd64 go build github.com/kif11/codercatapi

deploy:
	make build
	ssh kiko@ubkif "pm2 stop codercatapi"
	scp codercatapi kiko@ubkif:/home/kiko/codercatapi/
	ssh kiko@ubkif "pm2 restart codercatapi"
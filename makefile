install_dep:
	GO111MODULE=on go get
	GO111MODULE=on go mod vendor
build_linux:
	GO111MODULE=on goos=linux go build -o ./bin/linux/output
build_current_platform:
	GO111MODULE=on go build -o ./bin/windows/output
build_and_run:
	make build_current_platform && ./bin/windows/output
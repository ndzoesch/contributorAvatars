# Contributor Avatars

This is a **quick&dirty** script to get all profile URLS and pictures of contributors to a GitHub project.
These data is then compiled into a html file.

It is possible to exclude contributors you don't want to see in the final output.

# Usage

- Install [go](https://golang.org/)
- Clone this repo using `git clone https://github.com/ndzoesch/contributorAvatars`
- Rename `config.yaml.dist` to `config.yaml`
- Edit `config.yaml` so it fits your needs
- Execute `go mod tidy` in the folder of `main.go` 
- Compile & run or use `go run .` and 
# Install

## APT
echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless

## From source
go install github.com/bspippi1337/restless/cmd/restless@latest

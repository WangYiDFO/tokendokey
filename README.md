# Token Dokey

Token Dokey is a small tool that retrieves your access token from an OAuth service using a refresh token.

## Getting Started

### Example Command

#### Unix-like Systems (Linux or macOS)
Use the following command in your terminal:
```sh
GOOS=windows GOARCH=amd64 go build -o tokendokey.exe
```
#### Windows
Use the command prompte or PowerShell:
```sh
set GOOS=windows
set GOARCH=amd64
go build -o tokendokey.exe
```

### Usage
#### Step 1: Initialize OAuth Client
Run the following command to initialize your OAuth client config:
```sh
tokendokey.exe init yourclientname
```

This will prompt you for client_id, client_secret, and token_issue_url. Once completed, it will create a folder with three files:

1.config.json: Holds configuration details.
1.access-token.txt: Holds the access token.
1.refresh-token.txt: Holds the refresh token.

#### Step 2: Obtain a Valid Refresh Token or Offline Token
Use your preferred method to get a valid refresh token or offline token (as offline token is a type of refresh token).

#### Step 3: Retrieve a New Access Token
Run the following command to get a new access token:
```sh
tokendokey.exe get-token yourclientname
```
This command will load the refresh token from your folder and retrieve a new access token. If the access token is still valid, it won’t trigger a new call. Once done, access-token.txt will hold a valid token, which you can use to authenticate yourself.

## Contributing
Feel free to fork this project, submit issues, and send pull requests. Contributions are always welcome!

## License
This project is licensed under the MIT License
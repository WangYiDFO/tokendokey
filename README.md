# Token Dokey

Token Dokey is a small tool that enables you to log in to OAuth/OIDC services with *Device Code flow* and retrieve your access token using a refresh token/offline token.

## Getting Started

### Build Command
Build your command from this repo. 
Or simply download from the Releases of this repo.
Or download the pre-built binary file for below links:

 - For Linux: [tokendokey-linux-amd64-0.0.4.zip](https://github.com/user-attachments/files/17624154/tokendokey-linux-amd64-0.0.4.zip)
 - For Windows: [tokendokey-windows-amd64-0.0.4.zip](https://github.com/user-attachments/files/17624156/tokendokey-windows-amd64-0.0.4.zip)

For other OS, please use source code to build yourself.

#### Build on Unix-like Systems (Linux or macOS)
Use the following command in your terminal:
```sh
GOOS=linux GOARCH=amd64 go build -o tokendokey
GOOS=windows GOARCH=amd64 go build -o tokendokey.exe
```
#### Build on Windows
Use the command prompt or PowerShell:
```sh
set GOOS=windows
set GOARCH=amd64
go build -o tokendokey.exe
```

#### Tip on Windows and compile to Linux binary file
Use Git Bash shell from your VSCode, you can build for both Windows and Linux
```sh
GOOS=linux GOARCH=amd64 go build -o tokendokey
GOOS=windows GOARCH=amd64 go build -o tokendokey.exe
```

### Usage
Steps in order to get your token:
1. one time setting for each client
2. login to you client
3. once logged in, can get-token as long as session not expired.
4. for service/offline batch job, use offline_token is recommended.

####  Initialize OAuth Client
Run the following command to initialize your OAuth client config:
```sh
tokendokey.exe init -c=myclient
```

This will prompt you for client_id, client_secret, and OAuth/OIDC discovery URL. Once completed, it will create a folder in your home directory under .tokendokey/myclient with three files:

1. config.json: Holds configuration details.
2. access_token.txt: Holds the access token.
3. refresh_token.txt: Holds the refresh token.

####  Obtain a Valid Refresh Token or Offline Token
Run the following command to log in the user via Device Code flow:
```sh
tokendokey.exe login -c=myclient
```
If you need to get an offline token:
```sh
tokendokey.exe login -c=myclient -o
```

#### Retrieve a New Access Token
Run the following command to get a new access token:
```sh
tokendokey.exe get-token -c=myclient
```
This command will load the refresh token from your folder and retrieve a new access token. If the access token is still valid, it won’t trigger a new call. Once done, `access_token.txt` will hold a valid token, which you can use to authenticate yourself.

In case you need to force refresh a new access token, no matter if it is still valid:
```sh
tokendokey.exe get-token -c=myclient -f
```

#### Logout
Run the following command to log out and remove access and refresh tokens:
```sh
tokendokey.exe logout -c=myclient
```

#### List
Run the following command to list the clients for current user:
```sh
tokendokey.exe list
```
When a clientname is provided, will display the settings of this client.
```sh
tokendokey.exe list -c=myclient
```

#### Delete
Run the following command to delete settings of a client, by provide clientname:
```sh
tokendokey.exe delete -c=myclient
```

#### Export/Import Configuration
To export the configuration of a client, use the export command:
```sh
tokendokey.exe export -c=myclient
```
This will export the configuration of the client myclient to a file named tokendokey.key in the current directory. The exported configuration includes the client's configuration details, access token, and refresh token.


To import the configuration of a client, use the import command:
```sh
tokendokey.exe import -c=myclient
```
This will import the configuration from the tokendokey.key file in the current directory and store it in the client's configuration directory (~/.tokendokey/myclient).

Note: When importing a configuration, the tokendokey.key file must be present in the current working directory. If the file does not exist, the import command will fail.

Export Command Options

-c: Specify the client name to export the configuration for.

Import Command Options

-c: Specify the client name to import the configuration for.

## Contributing
Feel free to fork this project, submit issues, and send pull requests. Contributions are always welcome!

## License
This project is licensed under the MIT License

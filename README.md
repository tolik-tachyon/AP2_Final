## Setup 

```sh
$ git clone https://github.com/tolik-tachyon/AP2_Final.git
$ cd AP2_Final/
$ go build -o nob
```

## Usage

```sh
$ ./nob --help                    # list all commands
$ ./nob -l                        # list all services
$ ./nob -s <service-name>         # start service
$ ./nob -frontend                 # start frontend server
$ ./nob -s <service-name> -dotenv # reads env vars from `.env`
```

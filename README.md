# Standard File Server, Go Implementation

Golang implementation of the [Standard File](https://standardfile.org/) protocol.


### Running your own server
You can run your own Standard File server, and use it with any SF compatible client (like Standard Notes).
This allows you to have 100% control of your data.
This server implementation is built with Go and can be deployed in seconds.

##### You may require to add `/api` to the url of your server if you plan to use this server with https://standardnotes.org/

#### Getting started

**Requirements**

- Go 1.7+
- SQLite3 database

**Instructions**

1. Initialize project:

```
go get github.com/tectiv3/standardfile
go install github.com/tectiv3/standardfile
```

2. Start the server:

```
standardfile
```

3. Stop the server:

```
standardfile -stop
```
### Configuration options

- starting from 0.4.0 server can use json,toml or yaml configuration file, example standardfile.json provided in this repo

#### Customize port and database location
```
--port 8080
```
and
```
--db /var/lib/sf.db
```
default port is `8888` and database file named `sf.db` will be created in working directory

- with --socket option you can set server to listen on unix socket

#### Run the server in foreground:
- useful when running as systemd service.

```
standardfile -foreground
```

This will not daemonise the service, which might be handy if you want to handle that on some other level, like with init system, inside docker container, etc. 

To stop the service, kill the process or press `ctrl-C` if running in terminal.

#### Migrations
To perform migrations run `standardfile -migrate`

_Perform migration upon updating to v0.2.0_

#### Disable registration
To disable registration run with `standardfile -noreg`

#### Handle CORS automatically
Run with -cors flag to enable automatic cors handling (needed for standardnotes app for example).

### Deploying to a live server
I suggest putting it behind nginx or [caddy](https://caddyserver.com/) with https enabled location.
- nginx sample config
```
server {
    server_name sf.example.com;
    listen 80;
    return 301 https://$server_name$request_uri;
}

server {
    server_name sf.example.com;
    listen 443 ssl http2;

    ssl_certificate /etc/letsencrypt/live/sf.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sf.example.com/privkey.pem;

    include snippets/ssl-params.conf;

    location / {
	add_header Access-Control-Allow-Origin '*' always;
	add_header Access-Control-Allow-Credentials true always;
	add_header Access-Control-Allow-Headers 'authorization,content-type' always;
	add_header Access-Control-Allow-Methods 'GET, POST, PUT, PATCH, DELETE, OPTIONS' always;
	add_header Access-Control-Expose-Headers 'Access-Token, Client, UID' always;

	if ($request_method = OPTIONS ) {
		return 200;
	}

	proxy_set_header        Host $host;
	proxy_set_header        X-Real-IP $remote_addr;
	proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header        X-Forwarded-Proto $scheme;

	proxy_pass          http://localhost:8888;
	proxy_read_timeout  90;
    }
}
```
- caddy sample config
```
sf.example.com {
    gzip

    proxy / localhost:8888 {
        transparent
    }
}
```

### Optional Environment variables

**SECRET_KEY_BASE**

JWT secret key

## Contributing
Contributions are encouraged and welcome. Currently outstanding items:

- Test suite

## License

Licensed under MIT

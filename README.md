<a href="https://zerodha.tech"><img src="https://zerodha.tech/static/images/github-badge.svg" align="right" /></a>

# HTTP Script Executor

A simple HTTP server to execute scripts.

## Usecase

Executing a script on a remote instance by issuing a HTTP request. This eliminates the need to configure SSH.

## Config

A sample config can be seen in `config.sample.toml`.

```toml
addr = "127.0.0.1:7777"
script_folder = "./"
```

- `addr` is the address to bind to.
- `script_folder` is the folder where the server searches for the given script.

## API

A script can be executed by `POST` call to `http://addr/scriptname` and the arguments for the script can be given as a json array.

#### Example

To execute a `test.sh` script with `a, b` as arguments, this is the curl call.

```sh
curl --request POST \
  --url http://127.1:7777/test.sh \
  --header 'content-type: application/json' \
  --data '["a", "b"]'
```

- If the script is successfully executed, this returns a `200` with combined output of `stdout`, `stderr`.
- If there's an error executing the script, `500` is returned with the error.
- If the server is not able to find the script in the given `scripts_dir`, `404` is returned.

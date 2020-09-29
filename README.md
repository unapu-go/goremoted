# goremoted
Golang projects remote server bridge

## Config

Example file (`config.yaml`):

```yaml
    http_client_timeout: 5s
    server:
      listeners:
        - addr: ":4000"
          tlss:
            cert_file: server.crt
            key_file: server.pem
            generate:
              hosts:
                - example.com
    fallback:
      redirect_to: "PROTO://WWW_HOST/URI"
      redirect_status: 303
    hosts:
      example.com:
        project_page: "https://doc.HOST/%s"
        patterns:
          "/my_client/{project}":
            destinations:
              - "https://github.com/example/my_client__%s"
          "/{project}":
            destinations:
              # detect first found
              - "https://github.com/example-public/%s"
              - "https://github.com/example-private/%s"
```
# ngx_http_set_backend

This is a work in progress. The goal is to develop a nginx plugin in Go (by making mandatory C code call Go code). This module will mimic what github made with nginx as described in [this article](http://githubengineering.com/rearchitecting-github-pages/).

````
location / {
  set_backend $backend;
  proxy_pass http://$backend$request_uri;

}
````

### rest api

The container should be started with a defined `DOMAIN` environment variable. You can then access the api by using this URL `api.$DOMAIN`.

- Create/Update: `curl -H api.$DOMAIN/entries.json -X POST -d '{ "host": "some.host.com", "backend": "localhost:3000" }'`
- Index: `curl -H api.$DOMAIN/entries.json `
- Delete: `curl -H api.$DOMAIN/entries/<url_encoded_host>.json`

Note: you can use `-H 'Host: api.l.io'` if you don't want to setup DNS.

### Architecture

nginx worker processes use the `ngx_http_set_backend` module every time it gets a http request in a location that has the `set_bakckend` directive. `ngx_http_set_backend` call a Go `c-shared` library (using `dlopen` and `dlsym`, see why in "Issues encountered"). This library asks to the `backend` process through a unix socket which backend to use according to the given host.

The `backend` process uses a key value store ([boltdb](https://github.com/boltdb/bolt)) to map a host to a backend. This key value store will later be manageable through a simple API that will be served directly by nginx using the host **l.io**.

This implies that both nginx and the `backend` processes run.

### TODOs
- [ ] make a release build (container as small as possible, without go curl git ..., optimized for nginx + nginx built with releases options)
- [x] unix socket should be accessible by the nobody user
- [ ] logs: backend logs + nginx log on stdout / stderr so everything is handled by docker
- [x] REST api to add backend / delete a backend / list backend / update backend (using gorilla mux as the router)
- [x] start implementing the database (boltdb) that will, from a Host header, find the corresponding IP address
- [x] integration tests
- [ ] github pages for errors
- [x] possible to set the domaine name $DOMAIN and respond to api.$DOMAIN
- [ ] make / route document the API
- [ ] unikernel ?

### Hacking

`docker` must be installed and running

1. `make` - compile the module
2. `make test` - run integration tests

resources:
- http://blog.ralch.com/tutorial/golang-sharing-libraries/
- https://www.airpair.com/nginx/extending-nginx-tutorial
- http://www.nginxguts.com/2011/09/configuration-directives/#more-343
- https://github.com/openresty/set-misc-nginx-module

### Issues encountered

When calling a function from my shared library (written in Go), I sometime get locked forever on a futex during my request. To solve this I experimented **a lot**:
- tried to tweak my docker image in all possible ways
- used `ltrace` and `strace` to debug
- tried to use the dynamic module feature of nginx: https://www.nginx.com/blog/dynamic-modules-nginx-1-9-11/

I finally found [this issue](https://github.com/golang/go/issues/12873) on the Go github repository, that basically taught me that the Go runtime is loaded when the module is loaded by nginx and that Go library built with `buildmode=c-shared` should never get loaded before a `fork` (if the forked process intend to use the shared library). nginx workers are forked by the master process and they use the library. That was my problem. To solve it, and make the library calls work consistently I used `dlfcn` (`dlopen`, `dlsym`). This allow me to load dynamically the library in the workers (so after the `fork`). This probably has a performance impact, but I don't really care for now :)

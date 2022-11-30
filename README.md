# service-provider-integration-scm-file-retriever
[![Container build](https://github.com/redhat-appstudio/service-provider-integration-scm-file-retriever/actions/workflows/build.yaml/badge.svg)](https://github.com/redhat-appstudio/service-provider-integration-scm-file-retriever/actions/workflows/build.yaml)
[![codecov](https://codecov.io/gh/redhat-appstudio/service-provider-integration-scm-file-retriever/branch/main/graph/badge.svg?token=MiQMw3V0wG)](https://codecov.io/gh/redhat-appstudio/service-provider-integration-scm-file-retriever)
Library for downloading files from a source code management sites

### About

This repository contains a library for retrieving files from various source management systems using a repository and file paths as the primary form of input.

The main idea is to allow users to download files from a different SCM providers without the necessity of knowing their APIs and/or download endpoints,
as well as take care about the authentication.

### Usage

Import 

```
import (
  "github.com/redhat-appstudio/service-provider-integration-scm-file-retriever/gitfile"
)
```


The main function signature looks as follows:  

```
func getFileContents(ctx context.Context, namespace, repoUrl, filepath, ref string, callback func(ctx context.Context, url string)) (io.ReadCloser, error) 
```
It expects the user namespace name to perform AccessToken related operations, three file location parameters, from which repository URL and path to file are mandatory , and optional ref for the branch/tags.
Function type parameter is a callback used when user authentication is needed, that function will be called with the URL to OAuth service, on which user need to be redirected, and can be controlled using the context.

### URL and path formats
Repository URLs may or may not contain `.git` suffixes. Paths are usual `/a/b/filename` format. Optional `ref` may
contain commit id, tag or branch name.

### Supported SCM providers

 - GitHub



## Demo server application

For the preview and testing purposes, there is demo server application developed, which consists of API endpoint,
simple UI page and websocket connection mechanism. It's source code located under `server` module.

### Building demo server application 

Simplest way to build demo server app is to use docker based build. Simply run `docker build server -t <image_tag>` from the root of repository,
and demo application image will be built.

### Deploying demo server application

There is a bunch of helpful scripts located at `server/hack` which can be used for different deployment scenarios.
The general prerequisite for all deployment types is to have `SPI_GITHUB_CLIENT_ID` and `SPI_GITHUB_CLIENT_SECRET` environment variables to be set locally, containing
correct values from registered GitHub OAuth application. 

#### Deploying on Kubernetes
  ...in progress

#### Deploying on Openshift
If you want to use your custom server image, update the image name or tag at `server/config/default/kustomization.yaml` in `images` section first.

1. deploy SPI (https://github.com/redhat-appstudio/service-provider-integration-operator)
2. deploy scm-file-retriever-server `kubectl apply -k server/config/openshift`
3. server url will be available at `https://file-retriever-server-service-spi-system.<cluster-url>`

#### Enabling CORS for demo page

If it is planned to use test/demo page of the file retriever server, it's URL must be added
as the allowed origin on OAuth service via argument or environment variable. 
Example:
```
kubectl set env deployment/spi-oauth-service ALLOWEDORIGINS=https://console.dev.redhat.com,https://file-retriever-server-service-spi-system.<cluster-url>.com  -n spi-system
```
(see [OAuth service configuration parameters](https://github.com/redhat-appstudio/service-provider-integration-operator/blob/main/docs/ADMIN.md#oauth-service-configuration-parameters) for more details).


   
### Known peculiarities
The most common problem which may occur during file resolving, is that configured OAuth application is not approved to access
the particular repository. So, user must read GitHub OAuth authorization window carefully, and request permissions if needed.
There also can be some inconsistency of the OAuth scopes, which may lead to token matching fail.

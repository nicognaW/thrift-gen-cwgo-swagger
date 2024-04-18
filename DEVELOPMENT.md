# Development

This document is for developers / contributors only.

## Introduction

This project combines the OpenAPI compilers from [google/gnostic](github.com/google/gnostic) and the IDL annotation definitions from [cloudwego/hertz](https://github.com/cloudwego/hertz) to generate OpenAPI documents from IDL files.

## Getting started

Make sure you have go 1.22 installed, and

```bash
git clone https://github.com/cloudwego-contrib/thrift-gen-cwgo-swagger
cd thrift-gen-cwgo-swagger
go mod tidy
```

To run the plugin as the `thriftgo` plugin

```bash
go install .
thriftgo -g go -p cwgo-swagger testdata/psm.thrift
```

The implementation of the plugin is heavily inspired by [hz code generator](https://github.com/cloudwego/hertz/blob/171630c2490fa1f1dffa4ed11020ff7fd09ce8de/cmd/hz/thrift/ast.go) and [gnostic compiler](https://github.com/google/gnostic/tree/ad271d568b713ad381ad6751cd8b950eade78d98/cmd/protoc-gen-openapi/generator/generator.go).
Please make sure you understand all the code and processes of this two projects before you start contributing.

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

## Status

> - :construction: means what I'm working on
> - :memo: means I made decisions myself, and not sure if it's the best way to do it
> - :question: means I need help or suggestions

- [ ] functions (presenting HTTP APIs) to paths
  - [x] :memo: service name and function name to operation metadata
  - [x] HTTP method annotation to name and operation field key
  - [ ] first parameter to http request fields
    - [ ] :question: first parameter name / type to something?
    - [ ] :construction: path parameters
    - [ ] :construction: field annotations to corresponding fields in the request
      - [ ] api.raw_body
      - [ ] api.query
      - [ ] api.header
      - [ ] api.cookie
      - [ ] api.body
      - [ ] api.path
      - [ ] api.form
    - [x] untagged fields to query parameters
    - [ ] :construction: types to schemas
    - [ ] api.vd to schema validation
    - [x] optional / required fields
    - [x] default values in the schema
    - [ ] multi-default values in the schema
    - [x] reserved comments to description
    - [ ] :question: other annotations to tags or something?
  - [x] function type to responses
    - [x] function type struct to response schema
    - [ ] :question: something else to response
  - [ ] support extended service
- [ ] :memo: reference extraction
  - [ ] thrift typedef to schema reference
  - [ ] :construction: thrift struct to schema reference
  - [ ] thrift enum to schema reference
  - [ ] cull unused schemas
- [ ] :question: plugin inputs as flags or something?
- [ ] :question: default host

package main

import (
	"fmt"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/plugin"
	v3 "github.com/google/gnostic/openapiv3"
	"log"
	"net/http"
	"sort"
)

const (
	infoURL = "https://github.com/nicognaw"
)

type Configuration struct {
	Version *string
}

// OpenAPIv3Generator holds internal state needed to generate an OpenAPIv3 document for a thriftgo plugin request.
type OpenAPIv3Generator struct {
	conf Configuration

	req  *plugin.Request
	resp *plugin.Response
}

func (g OpenAPIv3Generator) Run() *plugin.Response {
	d := g.buildDocumentV3()
	bytes, err := d.YAMLValue("Generated with thrift-gen-openapi\n" + infoURL)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to build OpenAPIv3 document: %s", err.Error())
		log.Println(errorMsg)
		g.resp.Error = &errorMsg
		return g.resp
	}
	g.resp.Contents = append(g.resp.Contents, &plugin.Generated{
		Name:    &filename,
		Content: string(bytes),
	})
	return g.resp
}

// buildDocumentV3 builds an OpenAPIv3 document for a plugin request.
func (g OpenAPIv3Generator) buildDocumentV3() *v3.Document {
	d := &v3.Document{}

	d.Openapi = "3.0.3"
	goNSName := g.findGoNSName()
	d.Info = &v3.Info{
		Version:     "0.0.1",
		Title:       goNSName + " API", // TODO: Make the these configurable as flag or some kinda input
		Description: "The API for " + goNSName,
	}
	d.Paths = &v3.Paths{}

	g.addPathsToDocumentV3(d, g.req.AST.GetServices())
	d.Components = &v3.Components{}
	d.Components.Schemas = &v3.SchemasOrReferences{}
	// Sort the tags.
	{
		pairs := d.Tags
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Tags = pairs
	}
	// Sort the paths.
	{
		pairs := d.Paths.Path
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Paths.Path = pairs
	}
	// Sort the schemas.
	{
		pairs := d.Components.Schemas.AdditionalProperties
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Name < pairs[j].Name
		})
		d.Components.Schemas.AdditionalProperties = pairs
	}
	return d
}

func (g OpenAPIv3Generator) findGoNSName() string {
	namespaces := g.req.AST.GetNamespaces()
	for _, namespace := range namespaces {
		if namespace.Language == "go" {
			return namespace.GetName()
		}
	}
	g.resp.Warnings = append(g.resp.Warnings, "No go namespace found")
	return ""
}

func (g OpenAPIv3Generator) addPathsToDocumentV3(d *v3.Document, services []*parser.Service) {
	for _, service := range services {
		// TODO: support extended service
		serviceName := service.GetName()
		for _, function := range service.GetFunctions() {
			functionName := function.GetName()
			operationID := serviceName + "_" + functionName
			description := "API for " + serviceName + " service's " + functionName + " function"

			// TODO: support other hz IDL annotations than only HttpMethodAnnotation
			rs := getAnnotations(function.Annotations, HttpMethodAnnotations)
			if len(rs) == 0 {
				continue
			}
			httpAnnos := httpAnnotations{}
			for k, v := range rs {
				httpAnnos = append(httpAnnos, httpAnnotation{
					method: k,
					path:   v,
				})
			}
			// turn the map into a slice and sort it to make sure getting the results in the same order every time
			sort.Sort(httpAnnos)
			if len(httpAnnos) == 0 || len(httpAnnos[0].path) == 0 {
				continue
			}
			method, path := httpAnnos[0].method, httpAnnos[0].path[0]
			var argument *parser.Field
			if len(function.Arguments) == 1 {
				argument = function.Arguments[0]
			} else {
				if len(function.Arguments) > 1 {
					g.resp.Warnings = append(g.resp.Warnings, "Only one argument is supported for now, check https://github.com/cloudwego/hertz/blob/171630c2490fa1f1dffa4ed11020ff7fd09ce8de/cmd/hz/thrift/ast.go#L136 for updates")
				}
				argument = nil
			}

			op, path := g.buildOperationV3(
				d,
				operationID,
				serviceName,
				description,
				path,
				argument,
				function.FunctionType,
			)

			pathItem := &v3.NamedPathItem{
				Name:  path,
				Value: &v3.PathItem{},
			}

			switch method {
			case http.MethodGet:
				pathItem.Value.Get = op
			case http.MethodPost:
				pathItem.Value.Post = op
			case http.MethodPut:
				pathItem.Value.Put = op
			case http.MethodDelete:
				pathItem.Value.Delete = op
			case http.MethodOptions:
				pathItem.Value.Options = op
			case http.MethodHead:
				pathItem.Value.Head = op
			case http.MethodPatch:
				pathItem.Value.Patch = op
			case "ANY":
				pathItem.Value.Get = op
				pathItem.Value.Post = op
				pathItem.Value.Put = op
				pathItem.Value.Delete = op
				pathItem.Value.Options = op
				pathItem.Value.Head = op
				pathItem.Value.Patch = op
			}

			d.Paths.Path = append(d.Paths.Path, pathItem)
		}
	}
}

func (g OpenAPIv3Generator) buildOperationV3(
	d *v3.Document,
	operationID string,
	tagName string,
	description string,
	path string,
	argument *parser.Field,
	functionType *parser.Type,
) (*v3.Operation, string) {
	// Initialize the list of operation parameters.
	var parameters []*v3.ParameterOrReference

	// TODO: handle path params

	if argument != nil && argument.Type.Name != "" {
		argumentStruct := g.findStruct(argument.Type.Name)
		if argumentStruct != nil && argumentStruct.Fields != nil {
			// TODO: handle other type of arguments with hz IDL annotations
			fieldParams := g.buildQueryParamsV3(argumentStruct.Fields)
			parameters = append(parameters, fieldParams...)
		}
	}

	responses := &v3.Responses{}
	responseStruct := g.findStruct(functionType.Name)
	if responseStruct != nil && responseStruct.Fields != nil {
		responseProperties := &v3.Properties{
			AdditionalProperties: []*v3.NamedSchemaOrReference{},
		}
		for _, field := range responseStruct.Fields {
			typeName := field.Type.Name
			responseProperties.AdditionalProperties = append(responseProperties.AdditionalProperties, &v3.NamedSchemaOrReference{
				Name: field.Name,
				Value: &v3.SchemaOrReference{
					Oneof: &v3.SchemaOrReference_Schema{
						Schema: &v3.Schema{
							Title: field.Name,
							Type:  typeName,
						},
					},
				},
			})
		}
		responses = &v3.Responses{
			ResponseOrReference: []*v3.NamedResponseOrReference{
				{
					Name: "200",
					Value: &v3.ResponseOrReference{
						Oneof: &v3.ResponseOrReference_Response{
							Response: &v3.Response{
								Description: "OK",
								// TODO: add schema for response
								Content: &v3.MediaTypes{
									AdditionalProperties: []*v3.NamedMediaType{
										{
											Name: "application/json",
											Value: &v3.MediaType{
												Schema: &v3.SchemaOrReference{
													Oneof: &v3.SchemaOrReference_Schema{
														Schema: &v3.Schema{
															Title:       responseStruct.Name,
															Type:        "object",
															Description: responseStruct.ReservedComments,
															Properties:  responseProperties,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}

	op := &v3.Operation{
		Tags:        []string{tagName},
		Description: description,
		OperationId: operationID,
		Parameters:  parameters,
		Responses:   responses,
	}
	// TODO: handle default host
	return op, path
}

func (g OpenAPIv3Generator) findStruct(structName string) *parser.StructLike {
	for _, structLike := range g.req.AST.GetStructs() {
		if structLike.GetName() == structName {
			return structLike
		}
	}
	return nil
}

func (g OpenAPIv3Generator) buildQueryParamsV3(fields []*parser.Field) []*v3.ParameterOrReference {
	var parameters []*v3.ParameterOrReference
	for _, field := range fields {
		typeName := field.GetType().GetName()

		// TODO: fill in the schema for the field
		fieldSchema := &v3.SchemaOrReference{
			Oneof: &v3.SchemaOrReference_Schema{
				Schema: &v3.Schema{},
			},
		}
		hasDefault := field.Default != nil
		if hasDefault {
			// TODO: handle other types
			switch typeName {
			case "bool":
				defaultValue := *field.Default.TypedValue.Identifier == "true"
				fieldSchema.Oneof.(*v3.SchemaOrReference_Schema).Schema.Default = &v3.DefaultType{
					Oneof: &v3.DefaultType_Boolean{
						Boolean: defaultValue,
					},
				}
			}
		}
		parameters = append(parameters, &v3.ParameterOrReference{
			Oneof: &v3.ParameterOrReference_Parameter{
				Parameter: &v3.Parameter{
					Name:        field.GetName(),
					In:          "query",
					Description: field.GetReservedComments(),
					Required:    field.GetRequiredness() == parser.FieldType_Required,
					Schema:      fieldSchema,
				},
			},
		})

	}
	return parameters
}

func NewOpenAPIv3Generator(req *plugin.Request) *OpenAPIv3Generator {
	return &OpenAPIv3Generator{
		conf: Configuration{
			Version: &req.Version,
		},
		req:  req,
		resp: &plugin.Response{},
	}
}

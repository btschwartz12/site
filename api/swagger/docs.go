// Package swagger Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/pics/delete/{id}": {
            "delete": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Delete a picture",
                "tags": [
                    "pictures"
                ],
                "summary": "Delete a picture",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Picture ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/api/pics/upload": {
            "post": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Upload a picture",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pictures"
                ],
                "summary": "Upload a picture",
                "parameters": [
                    {
                        "type": "file",
                        "description": "Picture file",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Description of the picture",
                        "name": "description",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/repo.Picture"
                        }
                    }
                }
            }
        },
        "/api/visitors": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Get the visitors",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "visitors"
                ],
                "summary": "Get visitors",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/repo.Visitor"
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "repo.Picture": {
            "type": "object",
            "properties": {
                "description": {
                    "type": "string"
                },
                "extension": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "pit": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "repo.Visitor": {
            "type": "object",
            "properties": {
                "city": {
                    "type": "string"
                },
                "country": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "ip": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                },
                "pit": {
                    "type": "string"
                },
                "region": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "description": "Please provide a valid api token",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Your API Title",
	Description:      "Your API Description",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

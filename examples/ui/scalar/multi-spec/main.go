// Example: multi-spec — browse v1 (Swagger 2.0) and v2 (OpenAPI 3.1) from a single UI.
//
// Run from this directory:
//
//	go run .
//	open http://localhost:8082
package main

import (
	"log"
	"net/http"

	scalar "github.com/nopereta/go-api-docs/ui/scalar"
)

// ── spec (inlined) ──────────────────────────────────────────────────────────
var specV2 = []byte(openAPISpec)
var specV1 = []byte(swaggerSpec)

const openAPISpec = `
{
  "openapi": "3.1.0",
  "info": {
    "title": "Tasks API",
    "version": "2.0.0",
    "description": "A simple task-management API — used as the go-scalar basic example.",
    "contact": {
      "name": "Acme Platform",
      "url": "https://acme.example.com"
    },
    "license": {
      "name": "MIT"
    }
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local dev"
    },
    {
      "url": "https://api.acme.example.com/v2",
      "description": "Production"
    }
  ],
  "tags": [
    {
      "name": "tasks",
      "description": "Task operations"
    },
    {
      "name": "users",
      "description": "User operations"
    }
  ],
  "paths": {
    "/tasks": {
      "get": {
        "operationId": "listTasks",
        "summary": "List tasks",
        "tags": [
          "tasks"
        ],
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "schema": {
              "type": "string",
              "enum": [
                "open",
                "in_progress",
                "done"
              ]
            }
          },
          {
            "name": "assignee",
            "in": "query",
            "schema": {
              "type": "string",
              "format": "uuid"
            }
          },
          {
            "name": "limit",
            "in": "query",
            "schema": {
              "type": "integer",
              "default": 20,
              "maximum": 100
            }
          },
          {
            "name": "cursor",
            "in": "query",
            "description": "Pagination cursor returned by previous response.",
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Paginated list of tasks",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/TaskPage"
                }
              }
            }
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      },
      "post": {
        "operationId": "createTask",
        "summary": "Create a task",
        "tags": [
          "tasks"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/TaskInput"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Task created",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Task"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      }
    },
    "/tasks/{id}": {
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "required": true,
          "schema": {
            "type": "string",
            "format": "uuid"
          }
        }
      ],
      "get": {
        "operationId": "getTask",
        "summary": "Get a task",
        "tags": [
          "tasks"
        ],
        "responses": {
          "200": {
            "description": "Task found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Task"
                }
              }
            }
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      },
      "patch": {
        "operationId": "updateTask",
        "summary": "Update a task",
        "tags": [
          "tasks"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/TaskPatch"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Task updated",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Task"
                }
              }
            }
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      },
      "delete": {
        "operationId": "deleteTask",
        "summary": "Delete a task",
        "tags": [
          "tasks"
        ],
        "responses": {
          "204": {
            "description": "Task deleted"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      }
    },
    "/tasks/{id}/comments": {
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "required": true,
          "schema": {
            "type": "string",
            "format": "uuid"
          }
        }
      ],
      "get": {
        "operationId": "listComments",
        "summary": "List comments on a task",
        "tags": [
          "tasks"
        ],
        "responses": {
          "200": {
            "description": "List of comments",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Comment"
                  }
                }
              }
            }
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      },
      "post": {
        "operationId": "addComment",
        "summary": "Add a comment to a task",
        "tags": [
          "tasks"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/CommentInput"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Comment added",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Comment"
                }
              }
            }
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      }
    },
    "/users/me": {
      "get": {
        "operationId": "getMe",
        "summary": "Get current user",
        "tags": [
          "users"
        ],
        "responses": {
          "200": {
            "description": "Current authenticated user",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/User"
                }
              }
            }
          },
          "401": {
            "$ref": "#/components/responses/Unauthorized"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      }
    },
    "/users/{id}": {
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "required": true,
          "schema": {
            "type": "string",
            "format": "uuid"
          }
        }
      ],
      "get": {
        "operationId": "getUser",
        "summary": "Get a user by ID",
        "tags": [
          "users"
        ],
        "responses": {
          "200": {
            "description": "User found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/User"
                }
              }
            }
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          }
        },
        "security": [
          {
            "bearerAuth": []
          }
        ]
      }
    }
  },
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "Task": {
        "type": "object",
        "required": [
          "id",
          "title",
          "status",
          "createdAt"
        ],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "readOnly": true
          },
          "title": {
            "type": "string",
            "minLength": 1,
            "maxLength": 255
          },
          "description": {
            "type": "string"
          },
          "status": {
            "type": "string",
            "enum": [
              "open",
              "in_progress",
              "done"
            ]
          },
          "priority": {
            "type": "string",
            "enum": [
              "low",
              "medium",
              "high"
            ],
            "default": "medium"
          },
          "assignee": {
            "$ref": "#/components/schemas/UserRef"
          },
          "dueAt": {
            "type": "string",
            "format": "date-time",
            "nullable": true
          },
          "createdAt": {
            "type": "string",
            "format": "date-time",
            "readOnly": true
          },
          "updatedAt": {
            "type": "string",
            "format": "date-time",
            "readOnly": true
          }
        }
      },
      "TaskInput": {
        "type": "object",
        "required": [
          "title"
        ],
        "properties": {
          "title": {
            "type": "string",
            "minLength": 1,
            "maxLength": 255
          },
          "description": {
            "type": "string"
          },
          "priority": {
            "type": "string",
            "enum": [
              "low",
              "medium",
              "high"
            ],
            "default": "medium"
          },
          "assigneeId": {
            "type": "string",
            "format": "uuid"
          },
          "dueAt": {
            "type": "string",
            "format": "date-time",
            "nullable": true
          }
        }
      },
      "TaskPatch": {
        "type": "object",
        "properties": {
          "title": {
            "type": "string",
            "minLength": 1,
            "maxLength": 255
          },
          "description": {
            "type": "string"
          },
          "status": {
            "type": "string",
            "enum": [
              "open",
              "in_progress",
              "done"
            ]
          },
          "priority": {
            "type": "string",
            "enum": [
              "low",
              "medium",
              "high"
            ]
          },
          "assigneeId": {
            "type": "string",
            "format": "uuid",
            "nullable": true
          },
          "dueAt": {
            "type": "string",
            "format": "date-time",
            "nullable": true
          }
        }
      },
      "TaskPage": {
        "type": "object",
        "required": [
          "items"
        ],
        "properties": {
          "items": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/Task"
            }
          },
          "nextCursor": {
            "type": "string",
            "nullable": true
          }
        }
      },
      "Comment": {
        "type": "object",
        "required": [
          "id",
          "body",
          "author",
          "createdAt"
        ],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "readOnly": true
          },
          "body": {
            "type": "string",
            "minLength": 1
          },
          "author": {
            "$ref": "#/components/schemas/UserRef"
          },
          "createdAt": {
            "type": "string",
            "format": "date-time",
            "readOnly": true
          }
        }
      },
      "CommentInput": {
        "type": "object",
        "required": [
          "body"
        ],
        "properties": {
          "body": {
            "type": "string",
            "minLength": 1
          }
        }
      },
      "User": {
        "type": "object",
        "required": [
          "id",
          "name",
          "email"
        ],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid",
            "readOnly": true
          },
          "name": {
            "type": "string"
          },
          "email": {
            "type": "string",
            "format": "email"
          },
          "avatarURL": {
            "type": "string",
            "format": "uri",
            "nullable": true
          },
          "createdAt": {
            "type": "string",
            "format": "date-time",
            "readOnly": true
          }
        }
      },
      "UserRef": {
        "type": "object",
        "required": [
          "id",
          "name"
        ],
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid"
          },
          "name": {
            "type": "string"
          }
        }
      },
      "Error": {
        "type": "object",
        "required": [
          "code",
          "message"
        ],
        "properties": {
          "code": {
            "type": "string"
          },
          "message": {
            "type": "string"
          },
          "details": {
            "type": "object",
            "additionalProperties": true
          }
        }
      }
    },
    "responses": {
      "BadRequest": {
        "description": "Invalid request",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Error"
            }
          }
        }
      },
      "Unauthorized": {
        "description": "Missing or invalid credentials",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Error"
            }
          }
        }
      },
      "NotFound": {
        "description": "Resource not found",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/Error"
            }
          }
        }
      }
    }
  }
}
`

const swaggerSpec = `
{
  "swagger": "2.0",
  "info": {
    "title": "Tasks API",
    "version": "1.0.0",
    "description": "Legacy v1 of the Tasks API (Swagger 2.0). Superseded by v2 (OpenAPI 3.1).",
    "contact": {
      "name": "Acme Platform"
    },
    "license": {
      "name": "MIT"
    }
  },
  "host": "api.acme.example.com",
  "basePath": "/v1",
  "schemes": [
    "https",
    "http"
  ],
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "securityDefinitions": {
    "bearerAuth": {
      "type": "apiKey",
      "in": "header",
      "name": "Authorization",
      "description": "JWT bearer token — prefix with \u0060Bearer\u0060"
    }
  },
  "tags": [
    {
      "name": "tasks",
      "description": "Task operations"
    },
    {
      "name": "users",
      "description": "User operations"
    }
  ],
  "paths": {
    "/tasks": {
      "get": {
        "operationId": "listTasks",
        "summary": "List tasks",
        "tags": [
          "tasks"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "parameters": [
          {
            "name": "status",
            "in": "query",
            "type": "string",
            "enum": [
              "open",
              "in_progress",
              "done"
            ]
          },
          {
            "name": "assignee",
            "in": "query",
            "type": "string"
          },
          {
            "name": "page",
            "in": "query",
            "type": "integer",
            "default": 1
          },
          {
            "name": "limit",
            "in": "query",
            "type": "integer",
            "default": 20
          }
        ],
        "responses": {
          "200": {
            "description": "Paginated list of tasks",
            "schema": {
              "$ref": "#/definitions/TaskList"
            }
          },
          "401": {
            "description": "Unauthorized",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "post": {
        "operationId": "createTask",
        "summary": "Create a task",
        "tags": [
          "tasks"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/TaskInput"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "Task created",
            "schema": {
              "$ref": "#/definitions/Task"
            }
          },
          "400": {
            "description": "Bad request",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    },
    "/tasks/{id}": {
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "required": true,
          "type": "string"
        }
      ],
      "get": {
        "operationId": "getTask",
        "summary": "Get a task",
        "tags": [
          "tasks"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "responses": {
          "200": {
            "description": "Task found",
            "schema": {
              "$ref": "#/definitions/Task"
            }
          },
          "404": {
            "description": "Not found",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "put": {
        "operationId": "replaceTask",
        "summary": "Replace a task (full update)",
        "tags": [
          "tasks"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/TaskInput"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Task updated",
            "schema": {
              "$ref": "#/definitions/Task"
            }
          },
          "404": {
            "description": "Not found",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      },
      "delete": {
        "operationId": "deleteTask",
        "summary": "Delete a task",
        "tags": [
          "tasks"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "responses": {
          "204": {
            "description": "Task deleted"
          },
          "404": {
            "description": "Not found",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    },
    "/users/me": {
      "get": {
        "operationId": "getMe",
        "summary": "Get current user",
        "tags": [
          "users"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "responses": {
          "200": {
            "description": "Current user",
            "schema": {
              "$ref": "#/definitions/User"
            }
          },
          "401": {
            "description": "Unauthorized",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    },
    "/users/{id}": {
      "parameters": [
        {
          "name": "id",
          "in": "path",
          "required": true,
          "type": "string"
        }
      ],
      "get": {
        "operationId": "getUser",
        "summary": "Get a user by ID",
        "tags": [
          "users"
        ],
        "security": [
          {
            "bearerAuth": []
          }
        ],
        "responses": {
          "200": {
            "description": "User found",
            "schema": {
              "$ref": "#/definitions/User"
            }
          },
          "404": {
            "description": "Not found",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "Task": {
      "type": "object",
      "required": [
        "id",
        "title",
        "status"
      ],
      "properties": {
        "id": {
          "type": "string",
          "readOnly": true
        },
        "title": {
          "type": "string"
        },
        "description": {
          "type": "string"
        },
        "status": {
          "type": "string",
          "enum": [
            "open",
            "in_progress",
            "done"
          ]
        },
        "priority": {
          "type": "string",
          "enum": [
            "low",
            "medium",
            "high"
          ]
        },
        "assigneeId": {
          "type": "string"
        },
        "dueAt": {
          "type": "string",
          "format": "date-time"
        },
        "createdAt": {
          "type": "string",
          "format": "date-time",
          "readOnly": true
        },
        "updatedAt": {
          "type": "string",
          "format": "date-time",
          "readOnly": true
        }
      }
    },
    "TaskInput": {
      "type": "object",
      "required": [
        "title"
      ],
      "properties": {
        "title": {
          "type": "string"
        },
        "description": {
          "type": "string"
        },
        "priority": {
          "type": "string",
          "enum": [
            "low",
            "medium",
            "high"
          ],
          "default": "medium"
        },
        "assigneeId": {
          "type": "string"
        },
        "dueAt": {
          "type": "string",
          "format": "date-time"
        }
      }
    },
    "TaskList": {
      "type": "object",
      "properties": {
        "items": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/Task"
          }
        },
        "total": {
          "type": "integer"
        },
        "page": {
          "type": "integer"
        },
        "limit": {
          "type": "integer"
        }
      }
    },
    "User": {
      "type": "object",
      "required": [
        "id",
        "name",
        "email"
      ],
      "properties": {
        "id": {
          "type": "string",
          "readOnly": true
        },
        "name": {
          "type": "string"
        },
        "email": {
          "type": "string",
          "format": "email"
        },
        "avatarURL": {
          "type": "string"
        },
        "createdAt": {
          "type": "string",
          "format": "date-time",
          "readOnly": true
        }
      }
    },
    "Error": {
      "type": "object",
      "required": [
        "code",
        "message"
      ],
      "properties": {
        "code": {
          "type": "string"
        },
        "message": {
          "type": "string"
        }
      }
    }
  }
}
`

func main() {

	h, err := scalar.New(
		scalar.WithSources(
			scalar.Source{URL: "/v2/openapi.json", Title: "v2 (OpenAPI 3.1)", Default: true},
			scalar.Source{URL: "/v1/swagger.json", Title: "v1 (Swagger 2.0)"},
		),
		scalar.WithTheme(scalar.ThemeMoon),
		scalar.WithBranding(scalar.Branding{
			Title:    "Acme Corp",
			Subtitle: "Tasks API",
		}),
		scalar.With(scalar.DisableAgent, scalar.DisableMCP, scalar.DarkMode),
		scalar.WithBaseServerURL("https://api.acme.example.com"),
		scalar.WithPageTitle("Tasks API — all versions"),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	mux.HandleFunc("GET /v2/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specV2)
	})
	mux.HandleFunc("GET /v1/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(specV1)
	})

	log.Println("listening on http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

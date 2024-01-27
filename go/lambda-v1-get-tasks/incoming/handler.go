package incoming

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"root/go/lambda-shared"
	"root/go/lambda-v1-get-tasks/core"
	"root/go/lambda-v1-get-tasks/outgoing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/jmoiron/sqlx"
)

type Handler struct {
	Database    *sqlx.DB
	TaskService *core.TaskService
}

func NewHandler() (*Handler, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load aws default config: %w", err)
	}

	secret, err := shared.GetSecret(cfg, os.Getenv("DB_SECRET_ID"))
	if err != nil {
		return nil, fmt.Errorf("failed to get database secret: %w", err)
	}

	dbConfig, err := shared.NewDatabaseConfig(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create database config: %w", err)
	}

	database, err := shared.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	taskRepository := outgoing.NewTaskRepository(database)
	taskService := core.NewTaskService(taskRepository)

	return &Handler{
		Database:    database,
		TaskService: taskService,
	}, nil
}

func (handler *Handler) Invoke(ctx context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	taskEntities, err := handler.TaskService.GetTasks(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to get taskEntities: %w", err)
	}

	tasksResponse := NewTasksResponse(taskEntities)
	tasksResponseBody, err := json.Marshal(tasksResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{}, fmt.Errorf("failed to marshal taskEntities: %w", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(tasksResponseBody),
	}, nil
}

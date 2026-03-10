package main

import (
	"go.uber.org/fx"
)

// @title           Notification Service API
// @version         1.0
// @description     API documentation for the notification service of the School Tracking System.
// @host            localhost:8080
// @BasePath        /api/v1
func main() {
	fx.New(AppModule()).Run()
}

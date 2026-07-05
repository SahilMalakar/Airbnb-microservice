package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/router"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
)

// Config holds runtime configuration for the application.
type Config struct {
	Address string
}

// Application is the top-level struct wiring together config and storage
// dependencies needed to run the server.
type Application struct {
	Config Config
}

// NewConfig builds a Config using environment variables, falling back to
// some defaults when they aren't set.
func NewConfig() *Config {
	port := config.GetEnvString("PORT", "8080")

	return &Config{
		Address: ":" + port,
	}
}

// NewApplication constructs an Application from a valid Config.
func NewApplication(cfg *Config) *Application {
	return &Application{
		Config: *cfg,
	}
}

// RunServer sets up the database, wires all dependencies, and starts the
// HTTP server.
func (a *Application) RunServer() error {

	conn, err := config.LoadDb()
	if err != nil {
		fmt.Println("Error setting up database:", err)
		return err
	}

	defer conn.Close()

	// Wire dependencies: DB → Repository → Service → Controller → Router
	userRepo := db.NewUserRepository(conn)
	userService := service.NewUserService(userRepo)
	userController := handler.NewUserController(userService)
	userRouter := router.NewUserRouter(userController)
	roleRepo := db.NewRoleRepository(conn)
	roleService := service.NewRoleService(roleRepo)
	roleController := handler.NewRoleController(roleService)
	roleRouter := router.NewRoleRouter(roleController)
	permissionRepo := db.NewPermissionRepository(conn)
	permissionService := service.NewPermissionService(permissionRepo)
	permissionController := handler.NewPermissionController(permissionService)
	permissionRouter := router.NewPermissionRouter(permissionController)
	rolePermissionRepo := db.NewRolePermissionRepository(conn)
	rolePermissionService := service.NewRolePermissionService(rolePermissionRepo)
	rolePermissionController := handler.NewRolePermissionController(rolePermissionService)
	rolePermissionRouter := router.NewRolePermissionRouter(rolePermissionController)
	userRoleRepo := db.NewUserRoleRepository(conn)
	userRoleService := service.NewUserRoleService(userRoleRepo)
	userRoleController := handler.NewUserRoleController(userRoleService)
	userRoleRouter := router.NewUserRoleRouter(userRoleController)

	chi := router.SetUpRouter(userRouter, roleRouter, permissionRouter, rolePermissionRouter, userRoleRouter)

	server := &http.Server{
		Addr:         a.Config.Address,
		Handler:      chi,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("server started listening on", a.Config.Address)

	return server.ListenAndServe()
}

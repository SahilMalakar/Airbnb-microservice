package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/cache"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/router"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
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
	utils.MustLoadSecrets()

	conn, err := config.LoadDb()
	if err != nil {
		fmt.Println("Error setting up database:", err)
		return err
	}

	fmt.Println("Database connected successfully")
	defer conn.Close()

	redisClient, err := config.LoadRedis()
	if err != nil {
		fmt.Println("Error setting up redis :", err)
		return err
	}

	fmt.Println("Redis connected successfully")
	defer redisClient.Close()

	refreshTokenStore := cache.NewRefreshTokenStore(redisClient)

	// Wire dependencies: DB → Repository → Service → Controller → Router
	// Repositories that UserService depends on must be created first.
	roleRepo := db.NewRoleRepository(conn)
	permissionRepo := db.NewPermissionRepository(conn)
	rolePermissionRepo := db.NewRolePermissionRepository(conn)
	userRoleRepo := db.NewUserRoleRepository(conn)
	userRepo := db.NewUserRepository(conn)

	// UserService now needs roleRepo + userRoleRepo to resolve default role
	// and embed roles/permissions into JWTs at login/signup/refresh.
	userService := service.NewUserService(userRepo, userRoleRepo, roleRepo,refreshTokenStore)
	userController := handler.NewUserController(userService)
	userRouter := router.NewUserRouter(userController)

	roleService := service.NewRoleService(roleRepo)
	roleController := handler.NewRoleController(roleService)
	roleRouter := router.NewRoleRouter(roleController)

	permissionService := service.NewPermissionService(permissionRepo)
	permissionController := handler.NewPermissionController(permissionService)
	permissionRouter := router.NewPermissionRouter(permissionController)

	rolePermissionService := service.NewRolePermissionService(rolePermissionRepo)
	rolePermissionController := handler.NewRolePermissionController(rolePermissionService)
	rolePermissionRouter := router.NewRolePermissionRouter(rolePermissionController)

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

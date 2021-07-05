package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// parseConfig parses the configuration either from the same folder, or
// from an explicit path.
func parseConfig(customConfig *string) (*Config, error) {
	if customConfig == nil || len(*customConfig) == 0 {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	} else {
		viper.SetConfigFile(*customConfig)
	}

	var config Config

	// Try to read the configuration file.
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Default config.
	viper.SetDefault("server.port", 8088)

	// Parse the configuration into the config object.
	err := viper.Unmarshal(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func authHandler(isAdmin bool, configAuthToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userCookie := session.Get("user")
		xAuthToken := c.GetHeader("x-auth-token")

		if (userCookie == nil || userCookie == "") && xAuthToken != configAuthToken {
			c.AbortWithStatus(http.StatusForbidden)
		} else {
			if isAdmin && xAuthToken != configAuthToken {
				user := &UserStruct{}
				json.Unmarshal([]byte(userCookie.(string)), &user)

				// Prevent anyone who is not logged in to view this page.
				if user == nil || !user.IsAdmin {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}

			c.Next()
		}
	}
}

func main() {
	customConfig := flag.String("config", "", "The path to a custom config file")
	flag.Parse()

	db, err := ConnectDB()

	if err != nil {
		panic(err)
	}

	config, err := parseConfig(customConfig)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Instanciate all of the route handlers here.
	productHandlers := &ProductHandlers{DB: db}
	orderHandlers := &OrderHandlers{
		DB:     db,
		Config: config,
	}
	userHandlers := &UserHandlers{
		DB:     db,
		Config: config,
	}

	adminAuthToken := config.Admin.Token

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Apply the sessions middleware.
	store := cookie.NewStore([]byte(config.Secret))
	router.Use(sessions.Sessions("mysession", store))

	// Set the static/public path.
	router.Use(static.Serve("/", static.LocalFile("./public", false)))

	router.Any("/", func(c *gin.Context) {
		session := sessions.Default(c)
		userCookie := session.Get("user")

		if userCookie == nil || userCookie == "" {
			c.Redirect(http.StatusMovedPermanently, "/login")
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "POS",
			"admin": false,
		})
	})

	router.GET("/admin", authHandler(true, adminAuthToken), func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Admin",
			"admin": true,
		})
	})

	router.POST("/login", userHandlers.Login)
	router.GET("/login", userHandlers.LoginPage)

	router.GET("/api/orders/earnings", authHandler(true, adminAuthToken), orderHandlers.GetTotalEarnings)
	router.GET("/api/orders/earnigns/:day", authHandler(true, adminAuthToken), orderHandlers.EarningsPerDay)
	router.GET("/api/orders/totals/export", authHandler(true, adminAuthToken), orderHandlers.ExportTotalEarnings)
	router.GET("/api/orders", authHandler(true, adminAuthToken), orderHandlers.GetOrders)
	router.POST("/api/order", authHandler(false, adminAuthToken), orderHandlers.CreateOrder)
	router.GET("/api/order/:orderId", authHandler(false, adminAuthToken), orderHandlers.PrintOrder)
	router.GET("/api/order/:orderId/pub", orderHandlers.PublicOrder)
	router.GET("/api/order/:orderId/qrcode", authHandler(false, adminAuthToken), orderHandlers.OrderQRCode)
	router.DELETE("/api/order/:orderId", authHandler(true, adminAuthToken), orderHandlers.ToggleOrder)

	router.POST("/api/product", authHandler(true, adminAuthToken), productHandlers.CreateProduct)
	router.GET("/api/products", productHandlers.ListProducts)

	router.Run(":" + strconv.Itoa(config.Server.Port))
}

package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

type (
	shorten struct {
		URL      string `json:"url"`
		ShortUrl string `json:"shortUrl"`
	}
	Message struct {
		Message string `json:"message"`
	}
	errorMessage struct {
		Error Message `json:"error"`
	}
)

var (
	longurls = map[string]string{}
	seq      = 1
)

var ctx context.Context = context.Background()

//----------
// Handlers
//----------

func getStats(c echo.Context) error {
	keyword, _ := getKeywordFromUrl(c.QueryParam("url"))
	details := c.QueryParam("details")
	switch details {
	case "count":
		count, err := getHitCount(keyword, getConnection(ctx))
		if err != nil {
			return c.JSON(400, &errorMessage{
				Error: Message{
					Message: err.Error(),
				},
			})
		}
		return c.JSON(http.StatusOK, map[string]int{
			"count": count})
	case "full":
		hits, err := getHits(keyword, getConnection(ctx))
		if err != nil {
			return c.JSON(400, &errorMessage{
				Error: Message{
					Message: err.Error(),
				},
			})
		}
		return c.JSON(http.StatusOK, hits)
	default:
		return c.JSON(404, &errorMessage{
			Error: Message{
				Message: "Invalid details",
			},
		})
	}
	return nil
}

func createShort(c echo.Context) error {
	originalUrl := c.QueryParam("url")
	// validation
	if originalUrl == "" {
		return c.JSON(400, &errorMessage{
			Error: Message{
				Message: "Url Not Provided",
			},
		})
	}
	shortCode, err := createShortUrl(originalUrl, getConnection(ctx))

	if err != nil {
		return c.JSON(400, &errorMessage{
			Error: Message{
				Message: "Keyword already exists",
			},
		})
	}
	host := c.Request().Host
	return c.JSON(http.StatusOK, &shorten{
		URL:      originalUrl,
		ShortUrl: fmt.Sprintf("%s/%s", host, shortCode),
	})
}

func redirect(c echo.Context) error {
	keyword := c.Param("keyword")
	if originalUrl, err := getOriginalUrl(keyword, getConnection(ctx)); err == nil {
		_, err := registerHit(keyword, c.Request().UserAgent(), getConnection(ctx))
		if err != nil {
			fmt.Println(err)
		}
		return c.Redirect(302, originalUrl)
	}
	return c.JSON(404, &errorMessage{
		Error: Message{
			Message: "Url not found",
		},
	})
}

func shortenURL(c echo.Context) error {
	url := c.QueryParam("url")
	if url == "" {
		return c.JSON(400, &errorMessage{
			Error: Message{
				Message: "URL is empty",
			},
		})
	}
	return nil
}

func main() {
	conn := getConnection(ctx)
	defer conn.Close()
	err := migrate(conn)
	if err != nil {
		fmt.Println(err)
	}
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	// Routes
	// e.GET("/")
	e.GET("/:keyword", redirect)
	e.GET("/register", createShort)
	e.POST("/stats", getStats)
	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

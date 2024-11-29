package middleware

import (
	"net/http"
	"fmt"
	"database/sql"
	"github.com/go-sql-driver/mysql"

	"github.com/labstack/echo/v4"
	"go-store/db"

)


var conn *sql.DB

func AuthMiddleware(requiredRole int) echo.MiddlewareFunc {
	e := echo.New()

	dbCfg := mysql.Config{
		User:   "jnguyen1",
		Passwd: "Jdogjust2002!!",
		DBName: "jnguyen1",
	}

	var err error
	conn, err = sql.Open("mysql", dbCfg.FormatDSN())
	if err != nil {
		e.Logger.Fatal(err)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			fmt.Println("enters here 123")
			cookie, err := ctx.Cookie("user_email")
			if err != nil || cookie.Value == "" {
				fmt.Println("exit here 123")
				return ctx.Redirect(http.StatusFound, "/?error=Must log in first")
			}

			role, err := db.CheckRole(conn, cookie.Value)
			fmt.Println("yereed: ", cookie.Value)
			if err != nil {
				fmt.Println(err)
				fmt.Println("exit poo")
				return ctx.Redirect(http.StatusFound, "/?error=Must log in first")
			}

			if role < requiredRole {
				return ctx.Redirect(http.StatusFound, "/?error=You are not authorized for that page!")
			}

			return next(ctx)
		}
	}
}

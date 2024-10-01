package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	service "github.com/tiagods/auth/internal/domain/services"
	"github.com/tiagods/auth/internal/infra/utils"
	"net/http"
)

func Register(c echo.Context) error {
	return nil
}

func Login(c echo.Context) error {
	login := &request.Login{}

	if err := extractPresenter(c, login); err != nil {
		return err
	}
	for i, usr := range users {
		if usr.Username == login.Username &&
			usr.Password == login.Password {

			tokens, err := generateTokenPair(i, usr)
			if err != nil {
				return err
			}
			return c.JSON(http.StatusOK, tokens)
		}
	}

	return echo.ErrUnauthorized
}

func extractPresenter(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		c.Logger().Error(err)
		return utils.JSON(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(i); err != nil {
		c.Logger().Error(err)
		return utils.JSON(c, http.StatusBadRequest, err)
	}
	return nil
}

func RefreshToken(c echo.Context) error {
	tokenReq := &request.RefreshToken{}
	if err := extractPresenter(c, tokenReq); err != nil {
		return err
	}

	result, err := service.RefreshToken(tokenReq)
	if err != nil {
		return utils.JSON(c, 0, err)
	}
	return c.JSON(http.StatusCreated, result)
}

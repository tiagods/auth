package extractor

import (
	"github.com/labstack/echo/v4"
)

func Extractor(c echo.Context, i interface{}) error {
	defaultBinder := echo.DefaultBinder{}
	if err := defaultBinder.BindHeaders(c, i); err != nil {
		return err
	}
	if err := c.Bind(i); err != nil {
		return err
	}
	if err := c.Validate(i); err != nil {
		return err
	}
	return nil
}

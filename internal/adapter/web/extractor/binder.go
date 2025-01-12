package extractor

import (
	"github.com/labstack/echo/v4"
)

func ExtractPresenter(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return err
	}
	defaultBinder := echo.DefaultBinder{}
	if err := defaultBinder.BindHeaders(c, i); err != nil {
		return err
	}
	if err := c.Validate(i); err != nil {

		return err
	}
	return nil
}

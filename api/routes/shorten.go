package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"time"
)

// here json:url specifies how the struct should be serialized to and from a JSON field
type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemainig  int           `json:"rate_limit"`
	XRateLimitRest time.Duration `json:"rate_limit_reset`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	//body parser binds the request body to the struct
	if err := c.BodyParser(&body); err != nil {
		//maps the response to the JSON format
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse the JSON"})
	}

	// implement the rate limiting

	//check if the input is the acutal URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL submitted"})
	}

	//check for the domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "You donnot have access to this"})
	}
	//enforce https

	body.URL = helpers.EnforceHTTP(body.URL)
}

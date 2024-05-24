package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bhusal-rj/url-shortner/database"
	"github.com/bhusal-rj/url-shortner/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
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

	r2 := database.CreateClient(1)
	defer r2.Close()

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		//if the finding of key value is Nil that means that there is no record in the database
		//so update the record in the database
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		valInt, _ := strconv.Atoi(val)

		if valInt <= 0 {
			//it returns the remaining time to live of a key that has a timeout
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "Rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

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
	r2.Decr(database.Ctx, c.IP())

	return nil
}

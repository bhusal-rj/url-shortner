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
	"github.com/google/uuid"
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
	XRateLimitRest time.Duration `json:"rate_limit_reset"`
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

	//check the user sent link is used by other or not
	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]

	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "URL is already in use",
		})
	}

	defer r.Close()

	error := r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second)

	if error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	resp := response{
		URL:            body.URL,
		CustomShort:    "",
		Expiry:         body.Expiry,
		XRateRemainig:  10,
		XRateLimitRest: 30,
	}

	r2.Decr(database.Ctx, c.IP()).Result()

	val, _ = r2.Get(database.Ctx, c.IP()).Result()

	resp.XRateRemainig, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()

	resp.XRateLimitRest = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("Domain") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}

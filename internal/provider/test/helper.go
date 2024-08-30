package test

import (
	"math/rand"
	"os"
	"time"
)

const (
	charSetAlphaNum    = "abcdefghijklmnopqrstuvwxyz012346789"
	charSetAlpha       = "abcdefghijklmnopqrstuvwxyz"
	resourceNameLength = 10
)

func GenerateRandomResourceName() string {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)

	result := make([]byte, resourceNameLength)
	for i := 0; i < resourceNameLength; i++ {
		result[i] = charSetAlpha[randIntRange(rng, 0, len(charSetAlpha))]
	}
	return string(result)
}

func GenerateRandomEnvName() string {
	envPrefix := "altinity"
	if v := os.Getenv("ALTINITYCLOUD_TEST_ENV_PREFIX"); v != "" {
		envPrefix = v
	}

	return envPrefix + "-dummy-" + GenerateRandomResourceName()
}

// randIntRange returns a random integer between min (inclusive) and max (exclusive).
func randIntRange(rng *rand.Rand, minVal, maxVal int) int {
	return minVal + rng.Intn(maxVal-minVal)
}

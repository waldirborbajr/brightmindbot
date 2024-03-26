package lib

import "github.com/rs/zerolog/log"

func FailOnError(err error) {
	switch {
	case err != nil:
		log.Error().Msgf("ERROR: %v", err)
	}
}

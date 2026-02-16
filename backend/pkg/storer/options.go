package storer

import "github.com/rs/zerolog"

type Option func(*Storer)

func WithLogger(logger zerolog.Logger) Option {
	return func(s *Storer) {
		s.log = logger
	}
}

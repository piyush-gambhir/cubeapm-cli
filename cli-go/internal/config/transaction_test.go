package config

import (
	"fmt"
	"sync"
	"testing"
)

func TestUpdatePreservesConcurrentMutations(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	const workers = 24
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Update(func(cfg *Config) error {
				if cfg.Profiles == nil {
					cfg.Profiles = make(map[string]Profile)
				}
				cfg.Profiles[fmt.Sprintf("agent-%02d", i)] = Profile{}
				return nil
			})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Profiles) != workers {
		t.Fatalf("got %d profiles after concurrent updates, want %d", len(cfg.Profiles), workers)
	}
}

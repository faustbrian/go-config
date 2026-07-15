package config_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	config "github.com/faustbrian/go-config"
	"github.com/faustbrian/go-config/configtest"
)

func TestPlanLoadsDeterministicallyInParallel(t *testing.T) {
	t.Parallel()

	type settings struct {
		Name   string            `config:"name"`
		Labels map[string]string `config:"labels"`
	}
	source := configtest.NewSource(
		config.SourceInfo{Name: "parallel", Priority: config.PriorityExplicitFiles},
		config.Document{Tree: map[string]any{
			"name":   "worker",
			"labels": map[string]any{"region": "eu"},
		}},
	)
	plan := configtest.MustPlan(t, source)

	const loads = 128
	failures := make(chan error, loads)
	var group sync.WaitGroup
	group.Add(loads)
	for index := range loads {
		go func() {
			defer group.Done()
			snapshot, err := config.Load[settings](context.Background(), plan)
			if err != nil {
				failures <- err
				return
			}
			value := snapshot.Value()
			if value.Name != "worker" || value.Labels["region"] != "eu" {
				failures <- fmt.Errorf("load %d returned %#v", index, value)
				return
			}
			value.Labels["region"] = "mutated"
			if snapshot.Value().Labels["region"] != "eu" {
				failures <- fmt.Errorf("load %d exposed snapshot mutation", index)
			}
		}()
	}
	group.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
}

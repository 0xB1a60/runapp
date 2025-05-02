package apps

import (
	"encoding/json"
	"os"
	"path"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/0xB1a60/runapp/internal/common"
	"github.com/0xB1a60/runapp/internal/util"
)

func List() ([]App, error) {
	homeDir, err := util.HomeDirPath()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []App{}, nil
		}
		return nil, err
	}

	res := make([]App, 0, len(files))
	idx := make(map[string]App, len(files))

	var mu sync.Mutex
	var g errgroup.Group

	for _, f := range files {
		if f.IsDir() {
			g.Go(func() error {
				configContent, err := os.ReadFile(path.Join(homeDir, f.Name(), common.FileConfig))
				if err != nil {
					return err
				}

				var app App
				if err := json.Unmarshal(configContent, &app); err != nil {
					return err
				}
				app.checkAndCorrectStatus()

				mu.Lock()
				res = append(res, app)
				idx[app.Name] = app
				mu.Unlock()
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	sort.Slice(res, func(i, j int) bool {
		pi := common.AppStatusPriority[res[i].Status]
		pj := common.AppStatusPriority[res[j].Status]
		if pi == pj {
			return res[i].Name < res[j].Name // tie-break by Name
		}
		return pi < pj
	})

	return res, nil
}

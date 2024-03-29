package registry

import (
	"context"
	"errors"
	"github.com/meschbach/go-junk-bucket/pkg/files"
	"github.com/meschbach/plaid/internal/plaid/controllers/projectfile"
	"github.com/meschbach/plaid/resources"
	"github.com/meschbach/plaid/resources/operator"
	"path"
	"path/filepath"
)

type alpha1Interpreter struct {
	res *resources.Client
}

func (a *alpha1Interpreter) Create(ctx context.Context, which resources.Meta, spec AlphaV1Spec, bridgeState *operator.KindBridgeState) (*registry, AlphaV1Status, error) {
	reg := &registry{
		projects: map[string]resources.Meta{},
	}
	//does the file exist?
	var persistentFile Config
	if err := files.ParseJSONFile(spec.AbsoluteFilePath, &persistentFile); err != nil {
		return reg, AlphaV1Status{
			Problem: err.Error(),
		}, nil
	}
	//create projects
	var problems []error
	registryFileRelative := filepath.Dir(spec.AbsoluteFilePath)
	for name, dir := range persistentFile.Services {
		var wd string
		if path.IsAbs(dir) {
			wd = dir
		} else {
			wd = filepath.Join(registryFileRelative, dir)
		}

		ref := resources.Meta{
			Type: projectfile.Alpha1,
			Name: name,
		}
		reg.projects[name] = ref
		if err := a.res.Create(ctx, ref, projectfile.Alpha1Spec{
			WorkingDirectory: wd,
			ProjectFile:      "plaid.json",
		}); err != nil {
			problems = append(problems, err)
		}
	}
	var problem string
	if problems == nil {
		problem = ""
	} else {
		e := errors.Join(problems...)
		problem = e.Error()
	}

	return reg, AlphaV1Status{
		Problem: problem,
	}, nil
}

func (a *alpha1Interpreter) Update(ctx context.Context, which resources.Meta, rt *registry, s AlphaV1Spec) (AlphaV1Status, error) {
	return AlphaV1Status{}, errors.New("todo")
}

func (a *alpha1Interpreter) Delete(ctx context.Context, which resources.Meta, rt *registry) error {
	var problems []error
	for _, project := range rt.projects {
		_, err := a.res.Delete(ctx, project)
		problems = append(problems, err)
	}
	return errors.Join(problems...)
}

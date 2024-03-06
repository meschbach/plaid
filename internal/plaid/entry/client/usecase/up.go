package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/meschbach/go-junk-bucket/pkg/reactors"
	client2 "github.com/meschbach/plaid/client"
	"github.com/meschbach/plaid/client/up"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/project"
	"github.com/meschbach/plaid/internal/plaid/controllers/projectfile"
	"github.com/meschbach/plaid/internal/plaid/registry"
	"github.com/meschbach/plaid/ipc/grpc/logger"
	"github.com/meschbach/plaid/ipc/grpc/reswire/client"
	"github.com/meschbach/plaid/resources"
	"github.com/thejerf/suture/v4"
	"go.opentelemetry.io/otel/codes"
	"os"
	"path/filepath"
)

type UpOptions struct {
	ReportUpdates bool
	// DeleteOnCompletion will delete the created project on completion
	DeleteOnCompletion bool
}

func Up(ctx context.Context, daemon *client.Daemon, rt *client2.Runtime, opts UpOptions) error {
	//
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Join(errors.New("wd lookup failed"), err)
	}
	baseName := filepath.Base(pwd)

	//register drain
	registeredDrain, err := daemon.LoggerV1.RegisterDrain(ctx, &logger.RegisterDrainRequest{Name: baseName})
	if err != nil {
		return err
	}
	output := client2.NewTerminalLogger()
	pump := client2.PumpLogDrainEvents(daemon.LoggerV1, registeredDrain, output)
	daemon.Tree.Add(pump)

	client := daemon.Storage
	var deleteConfig *resources.Meta
	//Do we have a configuration file?
	if plaidConfigFile, has := os.LookupEnv("PLAID_CONFIG"); has {
		configRef := resources.Meta{
			Type: registry.AlphaV1,
			Name: baseName,
		}
		deleteConfig = &configRef
		if err := upCreateRegistry(ctx, client, configRef, plaidConfigFile); err != nil {
			return err
		}
	}

	name := baseName
	ref := resources.Meta{
		Type: projectfile.Alpha1,
		Name: name,
	}

	type waitCondition struct {
		met         bool
		taskSuccess bool
	}
	r, rin := reactors.NewChannel[*waitCondition](4)
	state := &waitCondition{met: false}

	var projectProgress *up.ReportProgress[project.Alpha1Status]
	var buildRun *up.ReportProgress[buildrun.AlphaStatus1]
	w, err := client.Watcher(ctx)
	if err != nil {
		return err
	}
	var token resources.WatchToken
	token, err = w.OnResource(ctx, ref, func(parent context.Context, changed resources.ResourceChanged) error {
		ctx, span := tracer.Start(parent, "onProjectChanged")
		defer span.End()

		var status projectfile.Alpha1Status
		exists, err := client.GetStatus(ctx, ref, &status)
		if err != nil {
			span.SetStatus(codes.Error, "failed to get status")
			if offErr := w.Off(ctx, token); offErr != nil {
				return errors.Join(offErr, err)
			}
			return err
		}
		if !exists {
			span.SetStatus(codes.Error, "missing")
			err := errors.New("gone missing")
			if offErr := w.Off(ctx, token); offErr != nil {
				return errors.Join(offErr, err)
			}
			return err
		}
		if opts.ReportUpdates {
			j, err := json.Marshal(status)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", j)
			if status.Current != nil {
				if projectProgress == nil || *status.Current == projectProgress.Of {
					projectProgress = &up.ReportProgress[project.Alpha1Status]{
						Prefix: fmt.Sprintf("<project %s>", (*status.Current).Name),
						Core:   daemon,
						Of:     *status.Current,
						OnChange: func(ctx context.Context, alpha1Status project.Alpha1Status) error {
							if len(alpha1Status.OneShots) == 1 && alpha1Status.OneShots[0].Ref != nil {
								buildRun = &up.ReportProgress[buildrun.AlphaStatus1]{
									Prefix: fmt.Sprintf("<oneshot %s>", alpha1Status.OneShots[0].Name),
									Core:   daemon,
									Of:     *alpha1Status.OneShots[0].Ref,
									OnChange: func(ctx context.Context, status buildrun.AlphaStatus1) error {
										return nil
									},
								}
								return buildRun.Watch(ctx)
							}
							return nil
						},
					}
					if err := projectProgress.Watch(ctx); err != nil {
						return err
					}
				} else {
					panic("todo: handle projectfile project change")
				}
			}
		}

		if status.Done {
			r.ScheduleStateFunc(ctx, func(ctx context.Context, state *waitCondition) error {
				if err := w.Off(ctx, token); err != nil {
					fmt.Println("[up-watcher] failed to turn off watch")
					return err
				}
				if err := w.Close(ctx); err != nil {
					fmt.Println("[up-watcher] failed to close watcher")
					return err
				}
				state.met = true
				state.taskSuccess = status.Success
				return nil
			})
		}
		return nil
	})
	if err != nil {
		return err
	}
	defer w.Close(ctx)

	if err := upCreateProject(ctx, client, pwd, "plaid.json", ref); err != nil {
		return err
	}

	for {
		select {
		case e := <-rin:
			if err := r.Tick(ctx, e, state); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}

		if state.met {
			if opts.DeleteOnCompletion {
				err := client.Delete(ctx, ref)
				if err != nil {
					fmt.Printf("Failed to delete project on completion because %s\n", err.Error())
				}

				if deleteConfig != nil {
					if err := client.Delete(ctx, *deleteConfig); err != nil {
						fmt.Printf("Failed to delete project registry because %s\n", err.Error())
					}
				}
			}
			if state.taskSuccess {
				rt.ExitCode = 0
			} else {
				rt.ExitCode = -1
			}
			return suture.ErrDoNotRestart
		}
	}
}

func upCreateRegistry(parent context.Context, client client.Client, configRef resources.Meta, plaidConfigFile string) error {
	ctx, span := tracer.Start(parent, "up.create-registry")
	defer span.End()

	spec := registry.AlphaV1Spec{AbsoluteFilePath: plaidConfigFile}
	if err := client.Create(ctx, configRef, spec); err != nil {
		return err
	}
	return nil
}

func upCreateProject(parent context.Context, client client.Client, baseDirectory, relativeProjectFile string, ref resources.Meta) error {
	ctx, span := tracer.Start(parent, "up.create-project")
	defer span.End()

	if err := client.Create(ctx, ref, projectfile.Alpha1Spec{
		WorkingDirectory: baseDirectory,
		ProjectFile:      relativeProjectFile,
	}); err != nil {
		return errors.Join(errors.New("failed to create"), err)
	}
	return nil
}

package project

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/meschbach/plaid/internal/junk"
	"github.com/meschbach/plaid/internal/plaid/controllers/buildrun"
	"github.com/meschbach/plaid/internal/plaid/controllers/exec"
	"github.com/meschbach/plaid/internal/plaid/controllers/service"
	"github.com/meschbach/plaid/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"sync"
	"testing"
	"time"
)

func TestProjectAlpha1(t *testing.T) {
	t.Run("Given a Plaid instance with the configured controller", func(t *testing.T) {
		onDone := junk.SetupTestTracing(t)
		t.Cleanup(func() {
			shutdown, done := context.WithTimeout(context.Background(), 1*time.Second)
			defer done()

			onDone(shutdown)
		})

		baseCtx, done := context.WithTimeout(context.Background(), 1*time.Second)
		t.Cleanup(done)
		ctx, _ := junk.TraceSubtest(t, baseCtx, tracer)
		plaid := resources.WithTestSubsystem(t, ctx)
		plaid.AttachController("plaid.controllers.project", NewProjectSystem(plaid.Controller))

		t.Run("When a new project is Created with a oneshot", func(t *testing.T) {
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			tmpDir := os.TempDir()
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			projectRef := resources.FakeMetaOf(Alpha1)
			projectSpec := Alpha1Spec{
				BaseDirectory: tmpDir,
				OneShots: []Alpha1OneShotSpec{
					{
						Name: "one-shot",
						Build: exec.TemplateAlpha1Spec{
							Command: "builder",
						},
						Run: exec.TemplateAlpha1Spec{
							Command: "run",
						},
					},
				},
			}
			require.NoError(t, plaid.Store.Create(ctx, projectRef, projectSpec))

			t.Run("Then a buildrun is Created", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var projectStatus Alpha1Status
				success, err := resources.WaitOn(ctx, resources.ForStatusState(plaid.Store, projectRef, &projectStatus, func(status *Alpha1Status) (bool, error) {
					if len(status.OneShots) < 1 {
						return false, nil
					}
					return status.OneShots[0].Ref != nil, nil
				}))
				require.NoError(t, err)
				assert.True(t, success, "successfully retrieved project status")

				var spec buildrun.AlphaSpec1
				exists, problem := plaid.Store.Get(ctx, *projectStatus.OneShots[0].Ref, &spec)
				require.NoError(t, problem)
				require.True(t, exists, "resource must still exist")

				t.Run("with the correct build configuration", func(t *testing.T) {
					assert.Equal(t, tmpDir, spec.Build.WorkingDir, "working directory set correctly")
					assert.Equal(t, "builder", spec.Build.Command, "builder command set correctly")
				})
				t.Run("with the correct run configuration", func(t *testing.T) {
					assert.Equal(t, tmpDir, spec.Run.WorkingDir, "working directory set correctly")
					assert.Equal(t, "run", spec.Run.Command, "run command set correctly")
				})
			})

			t.Run("And the buildrun exits for both build and run", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var buildRunRef *resources.Meta
				var buildRunStatus Alpha1Status
				success, err := resources.WaitOn(ctx, resources.ForStatusState(plaid.Store, projectRef, &buildRunStatus, func(status *Alpha1Status) (bool, error) {
					buildRunRef = status.OneShots[0].Ref
					return buildRunRef != nil, nil
				}))
				require.NoError(t, err)
				require.True(t, success)
				br := &BuildRunStatusMock{
					Ref:   buildRunRef,
					Store: plaid.Store,
				}
				require.NoError(t, br.FinishNow(ctx, 0))

				t.Run("Then the exit is noted in the project status", func(t *testing.T) {
					ctx, _ = junk.TraceSubtest(t, ctx, tracer)
					var status Alpha1Status
					exists, err := resources.WaitOn(ctx, resources.ForStatusState[Alpha1Status](plaid.Store, projectRef, &status, func(status *Alpha1Status) (bool, error) {
						return status.Done, nil
					}))
					require.NoError(t, err, "Error with status %#v", status)
					require.True(t, exists, "must exist")

					if assert.Len(t, status.OneShots, 1) {
						assert.True(t, status.OneShots[0].Done, "must have completed")
					}

					assert.True(t, status.Done, "project is completed.")
				})
			})
		})

		t.Run("When a new project is Created with a daemon service", func(t *testing.T) {
			tmpDir := os.TempDir()
			ctx, _ = junk.TraceSubtest(t, ctx, tracer)
			projectRef := resources.FakeMetaOf(Alpha1)
			projectWatcher, err := plaid.Store.Watcher(ctx)
			require.NoError(t, err)

			go func() {
				for {
					select {
					case e := <-projectWatcher.Feed:
						if err := projectWatcher.Digest(ctx, e); err != nil {
							panic(err)
						}
					case <-ctx.Done():
						return
						//todo: don't ignore
					}
				}
			}()

			builderCommand := faker.Word()
			runCommand := faker.Word()
			projectSpec := Alpha1Spec{
				BaseDirectory: tmpDir,
				Daemons: []Alpha1DaemonSpec{
					{
						Name: "daemon",
						Build: &exec.TemplateAlpha1Spec{
							Command: builderCommand,
						},
						Run: exec.TemplateAlpha1Spec{
							Command: runCommand,
						},
					},
				},
			}
			statusChange := newChangeTracker()
			_, err = projectWatcher.OnResource(ctx, projectRef, func(ctx context.Context, changed resources.ResourceChanged) error {
				switch changed.Operation {
				case resources.StatusUpdated:
					statusChange.Update()
				default:
				}
				return nil
			})
			require.NoError(t, err)

			createStatusChange := statusChange.Fork()
			require.NoError(t, plaid.Store.Create(ctx, projectRef, projectSpec))

			t.Run("Then the initial status is setup", func(t *testing.T) {
				createStatusChange.Wait()
				var status Alpha1Status
				exists, err := plaid.Store.GetStatus(ctx, projectRef, &status)
				require.NoError(t, err)
				assert.True(t, exists, "must exist")
				assert.False(t, status.Ready, "initial status must not be ready")
			})

			t.Run("Then a service is Created", func(t *testing.T) {
				ctx, _ = junk.TraceSubtest(t, ctx, tracer)
				var found []resources.Meta
				for len(found) == 0 {
					var err error
					found, err = plaid.Store.FindClaimedBy(ctx, projectRef, []resources.Type{service.Alpha1})
					require.NoError(t, err)

					time.Sleep(25 * time.Millisecond)
				}
				if assert.Len(t, found, 1) {
					var spec service.Alpha1Spec
					exists, problem := plaid.Store.Get(ctx, found[0], &spec)
					require.NoError(t, problem)
					require.True(t, exists, "resource must still exist")

					t.Run("with the correct build configuration", func(t *testing.T) {
						assert.Equal(t, tmpDir, spec.Build.WorkingDir, "working directory set correctly")
						assert.Equal(t, builderCommand, spec.Build.Command, "builder command set correctly")
					})
					t.Run("with the correct run configuration", func(t *testing.T) {
						assert.Equal(t, tmpDir, spec.Run.WorkingDir, "working directory set correctly")
						assert.Equal(t, runCommand, spec.Run.Command, "run command set correctly")
					})

					t.Run("Then the service is not ready", func(t *testing.T) {
						serviceChange := statusChange.Fork()
						exists, err := plaid.Store.UpdateStatus(ctx, found[0], service.Alpha1Status{
							Dependencies: nil,
							Build: service.Alpha1BuildStatus{
								State: Alpha1StateSuccess,
							},
							Ready: false,
						})
						require.NoError(t, err)
						assert.True(t, exists, "service must exist")

						serviceChange.Wait()
						var projectStatus Alpha1Status
						exists, err = plaid.Store.GetStatus(ctx, projectRef, &projectStatus)
						require.NoError(t, err)
						require.True(t, exists, "resource should exist")

						assert.False(t, projectStatus.Ready, "service is not ready")
					})

					t.Run("When the service is ready", func(t *testing.T) {
						serviceChange := statusChange.Fork()
						exists, err := plaid.Store.UpdateStatus(ctx, found[0], service.Alpha1Status{
							Dependencies: nil,
							Build: service.Alpha1BuildStatus{
								State: Alpha1StateSuccess,
							},
							Ready: true,
						})
						require.NoError(t, err)
						assert.True(t, exists, "service must exist")

						serviceChange.Wait()
						var projectStatus Alpha1Status
						exists, err = plaid.Store.GetStatus(ctx, projectRef, &projectStatus)
						require.NoError(t, err)
						require.True(t, exists, "resource should exist")

						assert.True(t, projectStatus.Ready, "service is ready")
					})
				}
			})
		})
	})
}

type changeTracker struct {
	lock   *sync.Mutex
	notice *sync.Cond
	epoch  uint
}

func newChangeTracker() *changeTracker {
	out := &changeTracker{
		lock:  &sync.Mutex{},
		epoch: 0,
	}
	out.notice = sync.NewCond(out.lock)
	return out
}

func (c *changeTracker) Fork() *changePoint {
	c.lock.Lock()
	defer c.lock.Unlock()
	return &changePoint{
		tracker:    c,
		afterEpoch: c.epoch,
	}
}

func (c *changeTracker) Update() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.epoch++
	c.notice.Broadcast()
}

type changePoint struct {
	tracker    *changeTracker
	afterEpoch uint
}

func (c *changePoint) Wait() {
	c.tracker.lock.Lock()
	defer c.tracker.lock.Unlock()
	for c.tracker.epoch <= c.afterEpoch {
		c.tracker.notice.Wait()
	}
}

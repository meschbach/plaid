package daemon

import (
	"context"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/buildrun"
	localexec "git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/exec/local"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/logdrain"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/project"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/projectfile"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/controllers/service"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/fileWatch"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/httpProbe"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/registry"
	"git.meschbach.com/mee/platform.git/plaid/internal/plaid/resources"
	"github.com/thejerf/suture/v4"
	"sync"
)

func newBundledService(resources *resources.Controller) *bundledServices {
	bundler := &bundledServices{
		Supervisor: suture.NewSimple("bundled-services"),
		resources:  resources,
		gate:       sync.WaitGroup{},
	}
	bundler.gate.Add(1)
	return bundler
}

type bundledServices struct {
	*suture.Supervisor
	resources *resources.Controller
	//todo: this is really wonky -- need to make it less clunky
	gate          sync.WaitGroup
	loggingConfig *logdrain.ServiceConfig
}

func (b *bundledServices) LogDrainConfig(ctx context.Context) *logdrain.ServiceConfig {
	b.gate.Wait()
	return b.loggingConfig
}

func (b *bundledServices) Serve(ctx context.Context) error {
	b.attachV0_1_0()
	b.attachV0_2_0()
	b.gate.Done()
	return b.Supervisor.Serve(ctx)
}

func (b *bundledServices) attachV0_2_0() {
	loggingSystem, loggingRoot := logdrain.NewLogDrainSystem(b.resources)
	b.loggingConfig = loggingSystem
	b.Supervisor.Add(loggingRoot)

	//todo: this should probably be sucked in since there are tie-ins with the controll and resources
	localProcSupervisors := suture.NewSimple("local-exec-procs")
	localExecSystem := localexec.NewSystem(b.resources, loggingSystem, localProcSupervisors)
	b.Supervisor.Add(localExecSystem)
	b.Supervisor.Add(localProcSupervisors)

	//todo: really devxp type stuff
	projectSystem := project.NewProjectSystem(b.resources)
	b.Supervisor.Add(projectSystem)
	localProjectFileSystem := projectfile.NewProjectFileSystem(b.resources)
	b.Supervisor.Add(localProjectFileSystem)
	b.Supervisor.Add(buildrun.NewSystem(b.resources))
	b.Supervisor.Add(service.NewSystem(b.resources))
}

func (b *bundledServices) attachV0_1_0() {
	// http probe tree
	probesTree := suture.NewSimple("probes/http")
	b.Add(probesTree)
	b.Add(httpProbe.NewController(b.resources, probesTree))
	//file watches
	fileWatchingSubtree := suture.NewSimple("files-watchers")
	fileWatch.NewFileWatch(fileWatchingSubtree, b.resources)
	b.Add(fileWatchingSubtree)

	//Legacy services which should be replaced
	b.Add(registry.NewController(b.resources))
}

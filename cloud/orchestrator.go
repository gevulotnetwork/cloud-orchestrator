package cloud

import (
	"errors"
	"log"
	"os"
	"path"

	"github.com/gevulotnetwork/cloud-orchestrator/config"
	"github.com/gevulotnetwork/cloud-orchestrator/fs"
	"github.com/gevulotnetwork/cloud-orchestrator/ops"
)

type Orchestrator struct {
	configFactory *config.Factory
}

func NewOrchestrator(cfgFactory *config.Factory) *Orchestrator {
	return &Orchestrator{
		configFactory: cfgFactory,
	}
}

func (o *Orchestrator) programToImageName(program string) string {
	// Bucket object name has maximum length of 63 and it requires first letter to
	// be a character so adjust the hash.
	imgNameLen := min(len(program), 62)
	return "a" + program[:imgNameLen]
}

func (o *Orchestrator) PrepareProgramImage(program string, baseImg string) error {
	program = o.programToImageName(program)

	// Tempdir needed for files from the original base image.
	tmp_dir, err := os.MkdirTemp(os.TempDir(), "gevulot")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp_dir)

	reader, err := fs.NewReader(baseImg)
	if err != nil {
		return err
	}

	err = copyAll(reader, tmp_dir, "/")
	if err != nil {
		return err
	}

	// Generate config for new image.
	cfg := o.configFactory.NewConfig(program)

	// Add files from the base image to new one.
	cfg.Dirs = append(cfg.Dirs, tmp_dir)

	// Copy program arguments
	cfg.Args = reader.ListArgs()

	// Assumption is that there is always at least the arg[0].
	if len(cfg.Args) > 0 {
		// Forming a program path this way is a best guess effort..
		cfg.Program = path.Join(tmp_dir, cfg.Args[0])
		cfg.ProgramPath = cfg.Program

		// Ensure that program is set executable.
		os.Chmod(cfg.Program, 0755)
	} else {
		return errors.New("base image doesn't have arguments; can't set config.Program")
	}

	// Create temporary file for the generated image.
	f, err := os.Create(path.Join(os.TempDir(), program))
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.RemoveAll(f.Name())

	// Copy env variables. Also configure the nanos kernel version if present.
	for k, v := range reader.ListEnv() {
		if cfg.Env == nil {
			cfg.Env = make(map[string]string)
		}

		cfg.Env[k] = v

		if k == "NANOS_VERSION" {
			cfg.NanosVersion = v
		}
	}

	cfg.CloudConfig.ImageName = program
	cfg.RunConfig.ImageName = f.Name()

	p, ctx, err := ops.Provider(cfg)
	if err != nil {
		return err
	}

	imgFile, err := p.BuildImage(ctx)
	if err != nil {
		return err
	}
	defer os.RemoveAll(imgFile)

	log.Printf("build new image: %q", imgFile)
	err = p.CreateImage(ctx, imgFile)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) CreateInstance(program string) (string, error) {
	program = o.programToImageName(program)

	cfg := o.configFactory.NewConfig(program)
	cfg.CloudConfig.ImageName = program
	cfg.RunConfig.InstanceName = program

	p, ctx, err := ops.Provider(cfg)
	if err != nil {
		return "", err
	}

	// TODO: Figure out if we actually need to pick up the right kernel version
	// from the image and configure it per program config?
	cfg.RunConfig.Kernel = cfg.Kernel

	log.Printf("creating instance on %s\n", cfg.CloudConfig.Platform)
	err = p.CreateInstance(ctx)
	if err != nil {
		return "", err
	}
	log.Printf("create instance %q on %s\n", cfg.RunConfig.InstanceName, cfg.CloudConfig.Platform)

	return cfg.RunConfig.InstanceName, nil
}

func (o *Orchestrator) DeleteInstance(program string) error {
	program = o.programToImageName(program)

	cfg := o.configFactory.NewConfig(program)
	cfg.CloudConfig.ImageName = program
	cfg.RunConfig.InstanceName = program

	p, ctx, err := ops.Provider(cfg)
	if err != nil {
		return err
	}

	log.Printf("deleting instance %q on %s\n", cfg.RunConfig.InstanceName, cfg.CloudConfig.Platform)
	err = p.DeleteInstance(ctx, program)
	if err != nil {
		return err
	}
	log.Printf("deleted instance %q on %s\n", cfg.RunConfig.InstanceName, cfg.CloudConfig.Platform)

	return nil
}

func copyAll(reader *fs.Reader, dstPath string, curPath string) error {
	fileEntries, err := reader.ReadDir(curPath)
	if err != nil {
		return err
	}

	for _, entry := range fileEntries {
		fullSrcPath := path.Join(curPath, entry.Name())
		fullDstPath := path.Join(dstPath, entry.Name())

		if entry.IsDir() {
			err = os.MkdirAll(fullDstPath, 0755)
			if err != nil {
				return err
			}

			// Recurse into sub-directories.
			err = copyAll(reader, fullDstPath, fullSrcPath)
			if err != nil {
				return err
			}

			continue
		}

		err = reader.CopyFile(fullSrcPath, fullDstPath, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

package glx_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/srvc/glx"
)

func TestNewConfig(t *testing.T) {
	yaml1 := `
<: &accounts-k8s
  name: accounts
  hostname: accounts.glx.local
  kubernetes:
    context: k8s.srvc.tools
    namespace: accounts
    labels:
      role: api
    ports:
      80: 8080

<: &admin-k8s
  name: admin
  hostname: admin.glx.local
  kubernetes:
    context: k8s.srvc.tools
    namespace: admin
    labels:
      role: web
    ports:
      80: 3000
`

	yaml2 := `
<: &blog-local
  name: blog
  hostname: blog.glx.local
  local:
    port_env:
      PORT: 80
    path: example.com/blog
    cmd: ["go", "run", "./cmd/server"]
`

	yaml3 := `
<: &admin-docker
  name: admin
  hostname: admin.glx.local
  docker:
    path: example.com/admin
    build:
      dockerfile: ./dev.dockerfile
    run:
      cmd: ["bin/rails", "s"]
      volumes:
      - .:/app
      - bundle:/usr/local/bundle
      - tmp:/app/tmp
      - log:/app/log
      port_envs:
      - PORT: 80
`

	yaml4 := `
root: /Users/glx/src

projects:
- name: blog
  apps:
  - *blog-local
  - *admin-k8s
  - *accounts-k8s

- name: admin
  apps:
  - *blog-local
  - *admin-docker
  - *accounts-k8s
`

	wantCfg := &glx.Config{
		Root: "/Users/glx/src",
		Projects: []*glx.Project{
			{
				Name: "blog",
				Apps: []*glx.App{
					{
						Name:     "blog",
						Hostname: "blog.glx.local",
						Local: &glx.LocalApp{
							PortEnv: map[string]glx.Port{"PORT": glx.Port(80)},
							Path:    "example.com/blog",
							Cmd:     []string{"go", "run", "./cmd/server"},
						},
					},
					{
						Name:     "admin",
						Hostname: "admin.glx.local",
						Kubernetes: &glx.KubernetesApp{
							Context:   "k8s.srvc.tools",
							Namespace: "admin",
							Labels:    map[string]string{"role": "web"},
							Ports:     map[glx.Port]glx.Port{80: 3000},
						},
					},
					{
						Name:     "accounts",
						Hostname: "accounts.glx.local",
						Kubernetes: &glx.KubernetesApp{
							Context:   "k8s.srvc.tools",
							Namespace: "accounts",
							Labels:    map[string]string{"role": "api"},
							Ports:     map[glx.Port]glx.Port{80: 8080},
						},
					},
				},
			},
			{
				Name: "admin",
				Apps: []*glx.App{
					{
						Name:     "blog",
						Hostname: "blog.glx.local",
						Local: &glx.LocalApp{
							PortEnv: map[string]glx.Port{"PORT": glx.Port(80)},
							Path:    "example.com/blog",
							Cmd:     []string{"go", "run", "./cmd/server"},
						},
					},
					{
						Name:     "admin",
						Hostname: "admin.glx.local",
						Docker: &glx.DockerApp{
							Ports: []glx.Port{80},
							Path:  "example.com/admin",
							Cmd:   []string{"bin/rails", "s"},
						},
					},
					{
						Name:     "accounts",
						Hostname: "accounts.glx.local",
						Kubernetes: &glx.KubernetesApp{
							Context:   "k8s.srvc.tools",
							Namespace: "accounts",
							Labels:    map[string]string{"role": "api"},
							Ports:     map[glx.Port]glx.Port{80: 8080},
						},
					},
				},
			},
		},
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "glx")

	cases := []struct {
		test  string
		setup func(t *testing.T, fs afero.Fs)
	}{
		{
			test: "with glx.yaml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "glx.k8s.yaml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.local.yaml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.docker.yaml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.yaml"), []byte(yaml4), 0644)
			},
		},
		{
			test: "without glx.yaml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "glx.k8s.yaml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.local.yaml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.docker.yaml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.projects.yaml"), []byte(yaml4), 0644)
			},
		},
		{
			test: "with glx.yml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "glx.k8s.yml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.local.yml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.docker.yml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.yml"), []byte(yaml4), 0644)
			},
		},
		{
			test: "without glx.yml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "glx.k8s.yml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.local.yml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.docker.yml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "glx.projects.yml"), []byte(yaml4), 0644)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.test, func(t *testing.T) {
			baseFs := afero.NewMemMapFs()
			tc.setup(t, baseFs)

			fs, err := glx.NewUnionFs(baseFs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			viper := glx.NewViper(fs)
			cfg, err := glx.NewConfig(viper)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(wantCfg, cfg); diff != "" {
				t.Errorf("loaded config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

package brokers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"csbbrokerpakaws/acceptance-tests/helpers/testpath"

	"csbbrokerpakaws/acceptance-tests/helpers/apps"
	"csbbrokerpakaws/acceptance-tests/helpers/cf"
	"csbbrokerpakaws/acceptance-tests/helpers/random"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type Option func(broker *Broker)

func CreateVm(opts ...Option) *Broker {
	broker := defaultVmConfig(opts...)

	out, err := exec.Command(
		"bosh",
		"int",
		"../assets/vars.yml",
		"-v", fmt.Sprintf("name=%s", broker.Name),
	).CombinedOutput()
	Expect(err).To(Succeed())
	Expect(os.WriteFile(fmt.Sprintf("vars-%s.yml", broker.Name), out, 0o644)).To(Succeed())

	deployCmd := exec.Command(
		"bosh",
		"-n", "deploy", "-d", broker.Name,
		"../assets/manifest.yml",
		"-l", fmt.Sprintf("vars-%s.yml", broker.Name),
		"-v", fmt.Sprintf("release_repo_path=%s", broker.boshReleaseDir),
		"-v", fmt.Sprintf("name=%s", broker.Name),
	)
	fmt.Println(deployCmd)

	session, err := gexec.Start(deployCmd, GinkgoWriter, GinkgoWriter)
	Expect(err).To(Not(HaveOccurred()))
	Eventually(session, time.Minute*30).Should(gexec.Exit(0))
	return &broker
}

func Create(opts ...Option) *Broker {
	broker := defaultConfig(opts...)

	brokerApp := apps.Push(
		apps.WithName(broker.Name),
		apps.WithDir(broker.dir),
		apps.WithManifest(newManifest(
			withName(broker.Name),
			withEnv(broker.env()...),
		)),
	)

	schemaName := strings.ReplaceAll(broker.Name, "-", "_")
	cf.Run("bind-service", broker.Name, "csb-sql", "-c", fmt.Sprintf(`{"schema":"%s"}`, schemaName))

	brokerApp.Start()

	cf.Run("create-service-broker", broker.Name, broker.username, broker.password, brokerApp.URL, "--space-scoped")

	broker.app = brokerApp
	return &broker
}

func WithOptions(opts ...Option) Option {
	return func(b *Broker) {
		for _, o := range opts {
			o(b)
		}
	}
}

func WithName(name string) Option {
	return func(b *Broker) {
		b.Name = name
	}
}

func WithVM() Option {
	return func(b *Broker) {
		b.isVmBased = true
	}
}

func WithPrefix(prefix string) Option {
	return func(b *Broker) {
		b.Name = random.Name(random.WithPrefix(prefix))
	}
}

func WithSourceDir(dir string) Option {
	Expect(filepath.Join(dir, "cloud-service-broker")).To(BeAnExistingFile())
	return func(b *Broker) {
		b.dir = dir
	}
}

func WithEnv(env ...apps.EnvVar) Option {
	return func(b *Broker) {
		b.envExtras = append(b.envExtras, env...)
	}
}

func WithBoshReleaseDir(dir string) Option {
	return func(b *Broker) {
		b.boshReleaseDir = dir
	}
}

func WithReleaseEnv(dir string) Option {
	return func(b *Broker) {
		b.envExtras = append(b.envExtras, readEnvrcServices(filepath.Join(dir, ".envrc"))...)
	}
}

func WithLatestEnv() Option {
	return func(b *Broker) {
		b.envExtras = append(b.envExtras, b.latestEnv()...)
	}
}

func WithUsername(username string) Option {
	return func(b *Broker) {
		b.username = username
	}
}

func WithPassword(password string) Option {
	return func(b *Broker) {
		b.password = password
	}
}

func defaultVmConfig(opts ...Option) (broker Broker) {
	defaults := []Option{
		WithName(random.Name(random.WithPrefix("broker"))),
		WithUsername(random.Name()),
		WithPassword(random.Password()),
		WithEncryptionSecret(random.Password()),
	}
	WithOptions(append(defaults, opts...)...)(&broker)
	return broker
}

func defaultConfig(opts ...Option) (broker Broker) {
	defaults := []Option{
		WithName(random.Name(random.WithPrefix("broker"))),
		WithSourceDir(testpath.BrokerpakRoot()),
		WithUsername(random.Name()),
		WithPassword(random.Password()),
		WithEncryptionSecret(random.Password()),
	}
	WithOptions(append(defaults, opts...)...)(&broker)
	return broker
}

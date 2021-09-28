package mole

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/davrodpin/mole/alias"
	"github.com/davrodpin/mole/fsutils"
	"github.com/davrodpin/mole/rpc"
	"github.com/davrodpin/mole/tunnel"

	"github.com/awnumar/memguard"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/mapstructure"
	daemon "github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	// IdFlagName is the name of the flag that carries the unique idenfier for a
	// mole instance.
	IdFlagName = "id"
)

// cli keeps a reference to the latest Client object created.
// This is mostly needed to introspect client states during runtime (e.g. a
// remote procedure call that needs to check certain runtime information)
var cli *Client

type Configuration struct {
	Id                string           `json:"id" mapstructure:"id" toml:"id"`
	TunnelType        string           `json:"tunnel-type" mapstructure:"tunnel-type" toml:"tunnel-type"`
	Verbose           bool             `json:"verbose" mapstructure:"verbose" toml:"verbose"`
	Insecure          bool             `json:"insecure" mapstructure:"insecure" toml:"insecure"`
	Detach            bool             `json:"detach" mapstructure:"detach" toml:"detach"`
	Source            AddressInputList `json:"source" mapstructure:"source" toml:"source"`
	Destination       AddressInputList `json:"destination" mapstructure:"destination" toml:"destination"`
	Server            AddressInput     `json:"server" mapstructure:"server" toml:"server"`
	Key               string           `json:"key" mapstructure:"key" toml:"key"`
	KeepAliveInterval time.Duration    `json:"keep-alive-interval" mapstructure:"keep-alive-interva" toml:"keep-alive-interval"`
	ConnectionRetries int              `json:"connection-retries" mapstructure:"connection-retries" toml:"connection-retries"`
	WaitAndRetry      time.Duration    `json:"wait-and-retry" mapstructure:"wait-and-retry" toml:"wait-and-retry"`
	SshAgent          string           `json:"ssh-agent" mapstructure:"ssh-agent" toml:"ssh-agent"`
	Timeout           time.Duration    `json:"timeout" mapstructure:"timeout" toml:"timeout"`
	SshConfig         string           `json:"ssh-config" mapstructure:"ssh-config" toml:"ssh-config"`
	Rpc               bool             `json:"rpc" mapstructure:"rpc" toml:"rpc"`
	RpcAddress        string           `json:"rpc-address" mapstructure:"rpc-address" toml:"rpc-address"`
}

// ParseAlias translates a Configuration object to an Alias object.
func (c Configuration) ParseAlias(name string) *alias.Alias {
	return &alias.Alias{
		Name:              name,
		TunnelType:        c.TunnelType,
		Verbose:           c.Verbose,
		Insecure:          c.Insecure,
		Detach:            c.Detach,
		Source:            c.Source.List(),
		Destination:       c.Destination.List(),
		Server:            c.Server.String(),
		Key:               c.Key,
		KeepAliveInterval: c.KeepAliveInterval.String(),
		ConnectionRetries: c.ConnectionRetries,
		WaitAndRetry:      c.WaitAndRetry.String(),
		SshAgent:          c.SshAgent,
		Timeout:           c.Timeout.String(),
		SshConfig:         c.SshConfig,
		Rpc:               c.Rpc,
		RpcAddress:        c.RpcAddress,
	}
}

// Client manages the overall state of the application based on its configuration.
type Client struct {
	Conf   *Configuration
	Tunnel *tunnel.Tunnel
	sigs   chan os.Signal
}

// New initializes a new mole's client.
func New(conf *Configuration) *Client {
	cli = &Client{
		Conf: conf,
		sigs: make(chan os.Signal, 1),
	}

	return cli
}

// Start kicks off mole's client, establishing the tunnel and its channels
// based on the client configuration attributes.
func (c *Client) Start() error {
	// memguard is used to securely keep sensitive information in memory.
	// This call makes sure all data will be destroy when the program exits.
	defer memguard.Purge()

	if c.Conf.Id == "" {
		u, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("could not auto generate app instance id: %v", err)
		}
		c.Conf.Id = u.String()[:8]
	}

	r, err := c.Running()
	if err != nil {
		log.WithFields(log.Fields{
			"id": c.Conf.Id,
		}).WithError(err).Error("error while checking for another instance using the same id")

		return err
	}

	if r {
		log.WithFields(log.Fields{
			"id": c.Conf.Id,
		}).Error("can't start. Another instance is already using the same id")

		return fmt.Errorf("can't start. Another instance is already using the same id %s", c.Conf.Id)
	}

	log.Infof("instance identifier is %s", c.Conf.Id)

	if c.Conf.Detach {
		var err error

		ic, err := NewDetachedInstance(c.Conf.Id)
		if err != nil {
			log.WithError(err).Errorf("error while creating directory to store mole instance related files")
			return err
		}

		err = startDaemonProcess(ic)
		if err != nil {
			log.WithFields(log.Fields{
				"id": c.Conf.Id,
			}).WithError(err).Error("error starting ssh tunnel")

			return err
		}
	} else {
		go c.handleSignals()
	}

	if c.Conf.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	d, err := fsutils.CreateInstanceDir(c.Conf.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"id": c.Conf.Id,
		}).WithError(err).Error("error creating directory for mole instance")

		return err
	}

	if c.Conf.Rpc {
		addr, err := rpc.Start(c.Conf.RpcAddress)
		if err != nil {
			return err
		}

		rd := filepath.Join(d.Dir, "rpc")

		err = ioutil.WriteFile(rd, []byte(addr.String()), 0644)
		if err != nil {
			log.WithFields(log.Fields{
				"id": c.Conf.Id,
			}).WithError(err).Error("error creating file with rpc address")

			return err
		}

		c.Conf.RpcAddress = addr.String()

		log.Infof("rpc server address saved on %s", rd)
	}

	t, err := createTunnel(c.Conf)
	if err != nil {
		log.WithFields(log.Fields{
			"id": c.Conf.Id,
		}).WithError(err).Error("error creating tunnel")

		return err
	}

	c.Tunnel = t

	if err = c.Tunnel.Start(); err != nil {
		log.WithFields(log.Fields{
			"tunnel": c.Tunnel.String(),
		}).WithError(err).Error("error while starting tunnel")

		return err
	}

	return nil
}

// Stop shuts down a detached mole's application instance.
func (c *Client) Stop() error {
	pfp, err := fsutils.GetPidFileLocation(c.Conf.Id)
	if err != nil {
		return fmt.Errorf("error getting information about aliases directory: %v", err)
	}

	if _, err := os.Stat(pfp); os.IsNotExist(err) {
		return fmt.Errorf("no instance of mole with id %s is running", c.Conf.Id)
	}

	cntxt := &daemon.Context{
		PidFileName: pfp,
	}

	d, err := cntxt.Search()
	if err != nil {
		return err
	}

	if c.Conf.Detach {
		err = os.RemoveAll(pfp)
		if err != nil {
			return err
		}
	} else {
		d, err := fsutils.InstanceDir(c.Conf.Id)
		if err != nil {
			return err
		}

		err = os.RemoveAll(d.Dir)
		if err != nil {
			return err
		}
	}

	err = d.Kill()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) handleSignals() {
	signal.Notify(c.sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	sig := <-c.sigs
	log.Debugf("process signal %s received", sig)
	c.Stop()
}

// Merge overwrites Configuration from the given Alias.
//
// Certain attributes like Verbose, Insecure and Detach will be overwritten
// only if they are found on the givenFlags which should contain the name of
// all flags given by the user through UI (e.g. CLI).
func (c *Configuration) Merge(al *alias.Alias, givenFlags []string) error {
	var fl flags

	fl = givenFlags

	if !fl.lookup("verbose") {
		c.Verbose = al.Verbose
	}

	if !fl.lookup("insecure") {
		c.Insecure = al.Insecure
	}

	if !fl.lookup("detach") {
		c.Detach = al.Detach
	}

	c.Id = al.Name
	c.TunnelType = al.TunnelType

	srcl := AddressInputList{}
	for _, src := range al.Source {
		err := srcl.Set(src)
		if err != nil {
			return err
		}
	}
	c.Source = srcl

	dstl := AddressInputList{}
	for _, dst := range al.Destination {
		err := dstl.Set(dst)
		if err != nil {
			return err
		}
	}
	c.Destination = dstl

	srv := AddressInput{}
	err := srv.Set(al.Server)
	if err != nil {
		return err
	}
	c.Server = srv

	c.Key = al.Key

	kai, err := time.ParseDuration(al.KeepAliveInterval)
	if err != nil {
		return err
	}
	c.KeepAliveInterval = kai

	c.ConnectionRetries = al.ConnectionRetries

	war, err := time.ParseDuration(al.WaitAndRetry)
	if err != nil {
		return err
	}
	c.WaitAndRetry = war

	c.SshAgent = al.SshAgent

	tim, err := time.ParseDuration(al.Timeout)
	if err != nil {
		return err
	}
	c.Timeout = tim

	c.SshConfig = al.SshConfig

	c.Rpc = al.Rpc

	c.RpcAddress = al.RpcAddress

	return nil
}

// ShowInstances returns the runtime information about all instances of mole
// running on the system with rpc enabled.
func ShowInstances() (*InstancesRuntime, error) {
	ctx := context.Background()
	data, err := rpc.ShowAll(ctx)
	if err != nil {
		return nil, err
	}

	var instances []Runtime

	err = mapstructure.Decode(data, &instances)
	if err != nil {
		return nil, err
	}

	runtime := InstancesRuntime(instances)

	if len(runtime) == 0 {
		return nil, fmt.Errorf("no instances were found.")
	}

	return &runtime, nil
}

// ShowInstance returns the runtime information about an application instance
// from the given id or alias.
func ShowInstance(id string) (*Runtime, error) {
	ctx := context.Background()
	info, err := rpc.Show(ctx, id)
	if err != nil {
		return nil, err
	}

	var r Runtime
	err = mapstructure.Decode(info, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func startDaemonProcess(instanceConf *DetachedInstance) error {
	args := appendIdArg(instanceConf.Id, os.Args)

	cntxt := &daemon.Context{
		PidFileName: instanceConf.PidFile,
		PidFilePerm: 0644,
		LogFileName: instanceConf.LogFile,
		LogFilePerm: 0640,
		Umask:       027,
		Args:        args,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		return err
	}

	if d != nil {
		err = os.Rename(instanceConf.PidFile, instanceConf.PidFile)
		if err != nil {
			return err
		}

		err = os.Rename(instanceConf.LogFile, instanceConf.LogFile)
		if err != nil {
			return err
		}

		log.Infof("execute \"mole stop %s\" if you like to stop it at any time", instanceConf.Id)

		os.Exit(0)
	}

	defer cntxt.Release()

	return nil
}

type flags []string

func (fs flags) lookup(flag string) bool {
	for _, f := range fs {
		if flag == f {
			return true
		}
	}

	return false
}

func createTunnel(conf *Configuration) (*tunnel.Tunnel, error) {
	s, err := tunnel.NewServer(conf.Server.User, conf.Server.Address(), conf.Key, conf.SshAgent, conf.SshConfig)
	if err != nil {
		log.Errorf("error processing server options: %v\n", err)
		return nil, err
	}

	s.Insecure = conf.Insecure
	s.Timeout = conf.Timeout

	err = s.Key.HandlePassphrase(func() ([]byte, error) {
		fmt.Printf("The key provided is secured by a password. Please provide it below:\n")
		fmt.Printf("Password: ")
		p, err := terminal.ReadPassword(int(syscall.Stdin))
		fmt.Printf("\n")
		return p, err
	})

	if err != nil {
		log.WithError(err).Error("error setting up password handling function")
		return nil, err
	}

	log.Debugf("server: %s", s)

	source := make([]string, len(conf.Source))
	for i, r := range conf.Source {
		source[i] = r.String()
	}

	destination := make([]string, len(conf.Destination))
	for i, r := range conf.Destination {
		if r.Port == "" {
			log.WithError(err).Errorf("missing port in destination address: %s", r.String())
			return nil, err
		}

		destination[i] = r.String()
	}

	t, err := tunnel.New(conf.TunnelType, s, source, destination, conf.SshConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	//TODO need to find a way to require the attributes below to be always set
	// since they are not optional (functionality will break if they are not
	// set and CLI parsing is the one setting the default values).
	// That could be done by make them required in the constructor's signature or
	// by creating a configuration struct for a tunnel object.
	t.ConnectionRetries = conf.ConnectionRetries
	t.WaitAndRetry = conf.WaitAndRetry
	t.KeepAliveInterval = conf.KeepAliveInterval

	return t, nil
}

// appendIdArg adds the id argument to the list of arguments passed by the user.
// This is helpful for scenarios where the process will be detached from the
// parent process and the new child process needs context about the instance.
func appendIdArg(id string, args []string) (newArgs []string) {
	for _, arg := range args {
		if arg == "--id" {
			return args
		}
	}

	newArgs = make([]string, len(args)+2)
	copy(newArgs, args)
	newArgs[len(args)-2] = fmt.Sprintf("--%s", IdFlagName)
	newArgs[len(args)-1] = id

	return
}

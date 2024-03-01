package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fasibio/portainer-api-cli/logger"
	"github.com/urfave/cli"
)

const (
	CliKeyDeployStack  = "deploystack"
	CliKeyUserName     = "username"
	CliKeyUserPassword = "password"
	CliStackName       = "stack"
	CliSwarmID         = "swarmid"
	CliComposePath     = "composepath"
	CliPortainerUrl    = "portainerurl"
	CliEndPointID      = "endpoint"
)

func getEnvName(prefix string) string {
	return fmt.Sprintf("PORTAINER_API_CLI_%s", strings.ToUpper(prefix))
}

func main() {
	r := Runner{}
	app := cli.NewApp()
	app.Name = "portainer-api-cli"
	app.Action = r.Run
	app.Before = r.Before
	app.Commands = []cli.Command{
		{
			Name: "create",
			Subcommands: cli.Commands{
				cli.Command{
					Name:   "config",
					Action: r.CreateConfig,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:   "name",
							EnvVar: getEnvName("name"),
							Usage:  "name of config",
						},
						cli.StringFlag{
							Name:   "content",
							EnvVar: getEnvName("content"),
							Usage:  "path to dir or - for STDIN",
						},
					},
				},
			},
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   CliKeyDeployStack,
			EnvVar: getEnvName(CliKeyDeployStack),
			Usage:  "Deploy stack to Portainer",
		},
		cli.StringFlag{
			Name:   CliKeyUserName,
			EnvVar: getEnvName(CliKeyUserName),
			Usage:  "username connect to portainer",
		},
		cli.StringFlag{
			Name:   CliKeyUserPassword,
			EnvVar: getEnvName(CliKeyUserPassword),
			Usage:  "password connect to portainer",
		},
		cli.StringFlag{
			Name:   CliStackName,
			EnvVar: getEnvName(CliStackName),
			Usage:  "name of stack you want deploy",
		},
		cli.StringFlag{
			Name:   CliSwarmID,
			EnvVar: getEnvName(CliSwarmID),
			Usage:  "the id of your swarm",
		},
		cli.StringFlag{
			Name:   CliComposePath,
			EnvVar: getEnvName(CliComposePath),
			Usage:  "set the elasticsearch destination index",
		},
		cli.StringFlag{
			Name:   CliPortainerUrl,
			EnvVar: getEnvName(CliPortainerUrl),
			Usage:  "Url to you portainer eg. http://portainer:9000",
		},
		cli.StringFlag{
			Name:   CliEndPointID,
			EnvVar: getEnvName(CliEndPointID),
			Usage:  "Endpoint to use",
			Value:  "1",
		},
	}
	if err := app.Run(os.Args); err != nil {
		logger.Get().Fatalw("Global error: " + err.Error())
	}
}

func getFileContent(path string) (string, error) {
	dat, err := os.ReadFile(path)
	return string(dat), err
}

type Runner struct {
	api PortainerApi
}

func (r *Runner) Before(c *cli.Context) error {
	logger.Initialize("info")
	r.api = PortainerApi{
		PortainerUrl: c.String(CliPortainerUrl),
	}
	err := r.api.Login(c.String(CliKeyUserName), c.String(CliKeyUserPassword))
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) CreateConfig(c *cli.Context) error {
	contentPath := c.String("content")
	content := ""
	if contentPath == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		content = ""
		for scanner.Scan() {
			content += scanner.Text() + "\r\n"
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	} else {
		c, err := getFileContent(contentPath)
		if err != nil {
			return err
		}
		content = c
	}
	log.Println("ioer", content)
	res, err := r.api.CreateConfig(c.String("name"), content, "1")
	if err == nil {
		logger.Get().Info(res)
	}
	return err
}

func (r *Runner) Run(c *cli.Context) error {
	logs := logger.Initialize("info")
	p := PortainerApi{
		PortainerUrl: c.String(CliPortainerUrl),
	}
	if c.Bool(CliKeyDeployStack) {
		logs.Info("Deploy Stack")
		err := p.Login(c.String(CliKeyUserName), c.String(CliKeyUserPassword))
		if err != nil {
			return err
		}
		composeContent, err := getFileContent(c.String(CliComposePath))
		if err != nil {
			return err
		}

		feedback, err := p.DeployNewApp(DeployNewStackInformation{
			Name:             c.String(CliStackName),
			SwarmID:          c.String(CliSwarmID),
			StackFileContent: composeContent,
		}, c.String(CliEndPointID))
		if err == nil {
			logs.Info("Deploy New Stack Successful ", feedback)
			return nil
		}
		logs.Info("Error Deploy new Stack perhaps it already exist... try to update", err)
		id, err := p.GetStackIDByName(c.String(CliStackName))
		if err != nil {
			return err
		}
		logs.Infow("Find Stack", "ID", id)
		feedback, err = p.UpdateStack(UpdateStackInfo{
			Prune:            false,
			StackFileContent: composeContent,
		}, id, c.String(CliEndPointID))
		if err != nil {
			return err
		}
		logs.Info("Update Stack Successful ", feedback)
		return nil
	}
	logs.Info("No Command choosed")

	return nil
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fasibio/portainer-api-cli/logger"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v2"
)

const (
	CliKeyUserName     = "username"
	CliKeyUserPassword = "password"
	CliStackName       = "name"
	CliSwarmID         = "swarmid"
	CliComposePath     = "composepath"
	CliPortainerUrl    = "portainerurl"
	CliEndPointID      = "endpoint"
)

func getEnvName(prefix string, path ...string) string {
	if len(path) == 0 {
		return fmt.Sprintf("PORTAINER_API_CLI_%s", strings.ToUpper(prefix))
	}
	p := strings.Join(path, "_")

	return fmt.Sprintf("PORTAINER_API_CLI_%s_%s", strings.ToUpper(p), strings.ToUpper(prefix))
}

func main() {
	r := Runner{}
	app := &cli.App{
		Name:   "portainer-api-cli",
		Before: r.Before,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    CliKeyUserName,
				EnvVars: []string{getEnvName(CliKeyUserName)},
				Usage:   "username connect to portainer",
			},
			&cli.StringFlag{
				Name:    CliKeyUserPassword,
				EnvVars: []string{getEnvName(CliKeyUserPassword)},
				Usage:   "password connect to portainer",
			},
			&cli.StringFlag{
				Name:    CliPortainerUrl,
				EnvVars: []string{getEnvName(CliPortainerUrl)},
				Usage:   "Url to you portainer eg. http://portainer:9000",
			},
			&cli.StringFlag{
				Name:    CliEndPointID,
				EnvVars: []string{getEnvName(CliEndPointID)},
				Usage:   "Endpoint to use",
				Value:   "1",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "config",
				Subcommands: []*cli.Command{
					{
						Name:   "rm",
						Action: r.ConfigRemove,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "id",
								EnvVars: []string{getEnvName("id", "config", "rm")},
							},
						},
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Action:  r.ConfigList,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "output",
								Aliases: []string{"o"},
								EnvVars: []string{getEnvName("output", "config", "ls")},
								Usage:   "output format of data, h for human readable or json",
								Value:   "h",
								Action: func(ctx *cli.Context, s string) error {
									if s == "h" || s == "json" {
										return nil
									}
									return fmt.Errorf("only h or json are allowed parameter")
								},
							},
						},
					},
					{
						Name:   "create",
						Usage:  "create a new swarm config",
						Action: r.ConfigCreate,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "name",
								EnvVars: []string{getEnvName("name", "config", "create")},
								Usage:   "name of config",
							},
							&cli.StringFlag{
								Name:    "content",
								EnvVars: []string{getEnvName("content", "config", "create")},
								Usage:   "path to dir or - for STDIN",
							},
							&cli.StringSliceFlag{
								Name:     "labels",
								EnvVars:  []string{getEnvName("labels", "config", "create")},
								Required: false,
							},
						},
					},
				},
			},
			{
				Name: "stack",
				Subcommands: []*cli.Command{
					{
						Name:   "deploy",
						Action: r.StackDeploy,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    CliStackName,
								EnvVars: []string{getEnvName(CliStackName, "stack", "deploy")},
								Usage:   "name of stack you want deploy",
							},
							&cli.StringFlag{
								Name:    CliSwarmID,
								EnvVars: []string{getEnvName(CliSwarmID, "stack", "deploy")},
								Usage:   "the id of your swarm",
							},
							&cli.StringFlag{
								Name:    CliComposePath,
								EnvVars: []string{getEnvName(CliComposePath, "stack", "deploy")},
								Usage:   "path to compose file",
							},
						},
					},
				},
			},
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
		EndpointId:   c.String(CliEndPointID),
	}
	err := r.api.Login(c.String(CliKeyUserName), c.String(CliKeyUserPassword))
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) ConfigRemove(c *cli.Context) error {
	return r.api.RemoveConfig(c.String("id"))
}

func (r *Runner) ConfigList(c *cli.Context) error {
	output := c.String("output")
	configs, err := r.api.ListConfig()
	if err != nil {
		return err
	}

	switch output {
	case "h":
		tbl := table.New("ID", "Name", "UpdatedAt")

		for _, v := range *configs {
			tbl.AddRow(v.ID, v.Spec.Name, v.UpdatedAt)
		}
		tbl.Print()
	case "json":
		res, err := json.Marshal(configs)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(res))
	}

	return nil
}

func (r *Runner) ConfigCreate(c *cli.Context) error {
	lables := c.StringSlice("labels")
	labelMap := make(map[string]string)
	for _, v := range lables {
		split := strings.SplitN(v, "=", 2)
		if len(split) != 2 {
			return fmt.Errorf("labels have to be key=value get %s %v", v, split)
		}
		labelMap[split[0]] = split[1]
	}

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
	res, err := r.api.CreateConfig(c.String("name"), content, labelMap)
	if err == nil {
		logger.Get().Info(res)
	}
	return err
}

func (r *Runner) StackDeploy(c *cli.Context) error {
	logs := logger.Initialize("info")
	p := PortainerApi{
		PortainerUrl: c.String(CliPortainerUrl),
	}
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

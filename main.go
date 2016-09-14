package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

var build = "0" // build number set at compile-time

func main() {
	//REMOVEME
	godotenv.Load("env-file")
	app := cli.NewApp()
	app.Name = "Drone google cloud storage plugin"
	app.Usage = "Drone google cloud storage plugin"
	app.Action = run
	app.Version = fmt.Sprintf("0.1.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "auth_key",
			Usage:  "Google cloud service account auth key",
			EnvVar: "GOOGLE_KEY",
		},
		cli.StringFlag{
			Name:  "source",
			Usage: "location of files to upload",
		},
		cli.StringFlag{
			Name:  "target",
			Usage: "destination to copy files to, including bucket name",
		},
		cli.StringFlag{
			Name:  "ignore",
			Usage: "kip files matching this pattern, relative to source",
		},
		cli.StringSliceFlag{
			Name:  "acl",
			Usage: "list of access rules applied to the uploaded files, in a form of entity:role",
		},
		cli.StringSliceFlag{
			Name:  "gzip",
			Usage: "files with the specified extensions will be gzipped and uploaded with \"gzip\" Content-Encoding header",
		},
		cli.StringFlag{
			Name:  "cache_control",
			Usage: "Cache-Control header",
		},
		cli.StringFlag{
			Name:  "metadata",
			Usage: "an arbitrary dictionary with custom metadata applied to all objects",
		},
		cli.StringFlag{
			Name:   "repo.owner",
			Usage:  "repository owner",
			EnvVar: "DRONE_REPO_OWNER",
		},
		cli.StringFlag{
			Name:   "repo.name",
			Usage:  "repository name",
			EnvVar: "DRONE_REPO_NAME",
		},
		cli.StringFlag{
			Name:   "commit.sha",
			Usage:  "git commit sha",
			EnvVar: "DRONE_COMMIT_SHA",
		},
		cli.StringFlag{
			Name:   "commit.branch",
			Value:  "master",
			Usage:  "git commit branch",
			EnvVar: "DRONE_COMMIT_BRANCH",
		},
		cli.StringFlag{
			Name:   "commit.author",
			Usage:  "git author name",
			EnvVar: "DRONE_COMMIT_AUTHOR",
		},
		cli.StringFlag{
			Name:   "commit.tag",
			Usage:  "commit tag",
			EnvVar: "DRONE_TAG",
		},
		cli.StringFlag{
			Name:   "build.event",
			Value:  "push",
			Usage:  "build event",
			EnvVar: "DRONE_BUILD_EVENT",
		},
		cli.IntFlag{
			Name:   "build.number",
			Usage:  "build number",
			EnvVar: "DRONE_BUILD_NUMBER",
		},
		cli.StringFlag{
			Name:   "build.status",
			Usage:  "build status",
			Value:  "success",
			EnvVar: "DRONE_BUILD_STATUS",
		},
		cli.StringFlag{
			Name:   "build.link",
			Usage:  "build link",
			EnvVar: "DRONE_BUILD_LINK",
		},
		cli.StringFlag{
			Name:  "env-file",
			Usage: "source env file",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.String("env-file") != "" {
		_ = godotenv.Load(c.String("env-file"))
	}

	plugin := Plugin{
		Repo: Repo{
			Owner: c.String("repo.owner"),
			Name:  c.String("repo.name"),
		},
		Build: Build{
			Number: c.Int("build.number"),
			Event:  c.String("build.event"),
			Status: c.String("build.status"),
			Link:   c.String("build.link"),
			Commit: c.String("commit.sha"),
			Branch: c.String("commit.branch"),
			Author: c.String("commit.author"),
			Tag:    c.String("commit.tag"),
		},
		Config: Config{
			AuthKey:      c.String("auth_key"),
			Source:       c.String("source"),
			Target:       c.String("target"),
			Ignore:       c.String("ignore"),
			Acl:          c.StringSlice("acl"),
			Gzip:         c.StringSlice("gzip"),
			CacheControl: c.String("cache_control"),
			Metadata:     c.String("metadata"),
		},
	}

	return plugin.Exec()
}

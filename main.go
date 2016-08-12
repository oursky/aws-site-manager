package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)
import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"gopkg.in/urfave/cli.v1"
)

func DisplayAwsErr(err error) {
	if err != nil {
		if strings.Contains(err.Error(), "security-credentials") {
			displayAwsCredentialHelp()
			os.Exit(1)
		}
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println("Error:", awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
			os.Exit(1)
		}
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func GetCallerReference() string {
	t := time.Now().Local()
	return t.Format("20060102150405")
}

func displayAwsCredentialHelp() {
	fmt.Println("Make sure credentials file setup: http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs")
	fmt.Println()
}

func checkDomain(c *cli.Context) {
	if !c.IsSet("domain") {
		fmt.Println("Domain was not set")
		cli.ShowCommandHelp(c, c.Command.Name)
		os.Exit(1)
	}
}

func main() {
	sess, err := session.NewSession()

	DisplayAwsErr(err)

	app := cli.NewApp()
	app.Name = "aws-site-manager"
	app.Usage = "Crete and sync static site with S3 and Cloudfront"
	app.Version = "0.1"
	app.Commands = []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create the S3 buckets and Cloudfront setting for a new website",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "domain",
					Value: "",
					Usage: "domain name of the site (without www)",
				},
				cli.BoolTFlag{
					Name:  "www",
					Usage: "add www for canonical domains, redirect www to naked domain",
				},
				cli.BoolTFlag{
					Name:  "www-noredirect",
					Usage: "add www for canonical domains, but don't redirect www to naked domain",
				},
				cli.BoolFlag{
					Name:  "ssl",
					Usage: "use ssl",
				},
				cli.StringFlag{
					Name:  "certBody",
					Usage: "path to PEM format certificate body",
				},
				cli.StringFlag{
					Name:  "certChain",
					Usage: "path to PEM format certificate chain",
				},
				cli.StringFlag{
					Name:  "privateKey",
					Usage: "path to PEM format private key",
				},
			},
			Action: func(c *cli.Context) error {
				checkDomain(c)
				certID := ""
				if c.Bool("ssl") {
					if !(c.IsSet("certBody") && c.IsSet("certChain") && c.IsSet("privateKey")) {
						cli.ShowCommandHelp(c, "create")
						os.Exit(1)
					}
					certID = UploadCert(c.String("domain"), c.String("certBody"), c.String("certChain"), c.String("privateKey"))
				}
				Create(sess, c.String("domain"), c.BoolT("www") || c.BoolT("www-noredirect"), certID)
				return nil
			},
		},
		{
			Name:    "sync",
			Aliases: []string{"s"},
			Usage:   "sync existing sites to S3 and invalidate cloudfront paths",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "domain",
					Value: "",
					Usage: "domain name of the site (without www)",
				},
				cli.BoolFlag{
					Name:  "reupload",
					Usage: "force an reupload",
				},
				cli.IntFlag{
					Name:  "concurrent",
					Usage: "number of concurrent upload to S3",
					Value: 4,
				},
				cli.StringFlag{
					Name:  "path",
					Usage: "path of files to upload",
					Value: ".",
				},
			},
			Action: func(c *cli.Context) error {
				checkDomain(c)
				Sync(sess, c.String("domain"), c.String("path"), c.Bool("reupload"), c.Int("concurrent"))
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelp(c)
		return nil
	}

	err = app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

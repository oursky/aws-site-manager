package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)
import (
	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/cloudfront"
	"github.com/awslabs/aws-sdk-go/service/iam"
	"github.com/awslabs/aws-sdk-go/service/s3"
	"github.com/codegangsta/cli"
)

var defaultRegion = "us-west-2"

func DisplayAwsErr(err error) {
	if awserr := aws.Error(err); awserr != nil {
		fmt.Println("Error:", awserr.Code, awserr.Message)
	} else if err != nil {
		panic(err)
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

func UploadCert(domain string, certBodyPath string, certChainPath string, privateKeyPath string) string {
	svc := iam.New(nil)

	certBody, err := ioutil.ReadFile(certBodyPath)
	CheckErr(err)
	certChain, err := ioutil.ReadFile(certChainPath)
	CheckErr(err)
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	CheckErr(err)

	uploadCertInput := &iam.UploadServerCertificateInput{
		CertificateBody:       aws.String(string(certBody)),
		CertificateChain:      aws.String(string(certChain)),
		PrivateKey:            aws.String(string(privateKey)),
		ServerCertificateName: aws.String(domain),
		Path: aws.String("/cloudfront/production/"),
	}

	resp, err := svc.UploadServerCertificate(uploadCertInput)

	if err != nil {
		panic(err)
	} else {
		fmt.Println(resp.ServerCertificateMetadata.ServerCertificateID)
		return *resp.ServerCertificateMetadata.ServerCertificateID
	}
}

func Create(domain string, www bool, certID string) {
	svc := s3.New(&aws.Config{Region: defaultRegion})

	// What it does:
	// Step 1: Create new bucket with domain name
	// Step 2: Enable Public Website Input

	bucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(domain), // Required
	}

	websiteInput := &s3.PutBucketWebsiteInput{
		Bucket: aws.String(domain),
		WebsiteConfiguration: &s3.WebsiteConfiguration{ // Required
			ErrorDocument: &s3.ErrorDocument{
				Key: aws.String("error.html"), // Required
			},
			IndexDocument: &s3.IndexDocument{
				Suffix: aws.String("index.html"), // Required
			},
		},
	}

	aliases := []*string{aws.String(domain)}
	quantityOfAliases := 1
	if www == true {
		aliases = append(aliases, aws.String("www."+domain))
		quantityOfAliases = 2
	}

	viewerCertificate := &cloudfront.ViewerCertificate{
		IAMCertificateID: aws.String(certID),
		SSLSupportMethod: aws.String("sni-only"),
	}

	distributionInput := &cloudfront.CreateDistributionInput{
		DistributionConfig: &cloudfront.DistributionConfig{
			Aliases: &cloudfront.Aliases{
				Quantity: aws.Long(int64(quantityOfAliases)),
				Items:    aliases,
			},
			CallerReference:   aws.String(GetCallerReference()),
			Comment:           aws.String(""),
			DefaultRootObject: aws.String("index.html"),
			DefaultCacheBehavior: &cloudfront.DefaultCacheBehavior{
				ForwardedValues: &cloudfront.ForwardedValues{
					Cookies: &cloudfront.CookiePreference{
						Forward: aws.String("none"),
					},
					QueryString: aws.Boolean(false),
				},
				MinTTL:         aws.Long(60),
				TargetOriginID: aws.String("S3-" + domain + "-SITE"),
				TrustedSigners: &cloudfront.TrustedSigners{
					Enabled:  aws.Boolean(false),
					Quantity: aws.Long(0),
				},
				ViewerProtocolPolicy: aws.String("allow-all"),
			},
			Enabled: aws.Boolean(true),
			Origins: &cloudfront.Origins{
				Quantity: aws.Long(1),
				Items: []*cloudfront.Origin{
					&cloudfront.Origin{
						DomainName: aws.String(domain + ".s3.amazonaws.com"),
						ID:         aws.String("S3-" + domain + "-SITE"),
						S3OriginConfig: &cloudfront.S3OriginConfig{
							OriginAccessIdentity: aws.String(""),
						},
					},
				},
			},
		},
	}

	if certID != "" {
		distributionInput.DistributionConfig.ViewerCertificate = viewerCertificate
	}

	_, err := svc.CreateBucket(bucketInput)
	DisplayAwsErr(err)

	_, err = svc.PutBucketWebsite(websiteInput)
	DisplayAwsErr(err)

	cf := cloudfront.New(nil)

	_, err = cf.CreateDistribution(distributionInput)
	DisplayAwsErr(err)
}

func Error() {
	fmt.Println("Usage: " + os.Args[0] + " [create|sync]")
	fmt.Println("Make sure credentials file setup: http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs")
	os.Exit(1)
func checkDomain(c *cli.Context) {
	if !c.IsSet("domain") {
		fmt.Println("Domain was not set")
		cli.ShowCommandHelp(c, c.Command.Name)
		os.Exit(1)
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "aws-site-manager"
	app.Usage = "Crete and sync static site with S3 and Cloudfront"
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
					Usage: "add www for canonical domains",
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
			Action: func(c *cli.Context) {
				checkDomain(c)
				certID := ""
				if c.Bool("ssl") {
					if !(c.IsSet("certBody") && c.IsSet("certChain") && c.IsSet("privateKey")) {
						cli.ShowCommandHelp(c, "create")
						os.Exit(1)
					}
					certID = UploadCert(c.String("domain"), c.String("certBody"), c.String("certChain"), c.String("privateKey"))
				}
				Create(c.String("domain"), c.BoolT("www"), certID)
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
			Action: func(c *cli.Context) {
				checkDomain(c)
				Sync(c.String("domain"), c.String("path"), c.Bool("reupload"), c.Int("concurrent"))
			},
		},
	}
	app.Action = func(c *cli.Context) {
		cli.ShowAppHelp(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

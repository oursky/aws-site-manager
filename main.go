package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)
import (
	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/cloudfront"
	"github.com/awslabs/aws-sdk-go/service/s3"
)

func DisplayAwsErr(err error) {
	if awserr := aws.Error(err); awserr != nil {
		fmt.Println("Error:", awserr.Code, awserr.Message)
	} else if err != nil {
		panic(err)
	}
}

func GetCallerReference() string {
	t := time.Now().Local()
	return t.Format("20060102150405")
}

func Create(domain string) {
	svc := s3.New(&aws.Config{Region: "us-west-2"})

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

	distributionInput := &cloudfront.CreateDistributionInput{
		DistributionConfig: &cloudfront.DistributionConfig{
			Aliases: &cloudfront.Aliases{
				Quantity: aws.Long(2),
				Items: []*string{
					aws.String(domain),
					aws.String("www." + domain),
				},
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

	_, err := svc.CreateBucket(bucketInput)
	DisplayAwsErr(err)

	_, err = svc.PutBucketWebsite(websiteInput)
	DisplayAwsErr(err)

	cf := cloudfront.New(nil)

	_, err = cf.CreateDistribution(distributionInput)
	DisplayAwsErr(err)
}

func Sync() {
	fmt.Println("TODO: Sync not implemented")
}

func Error() {
	fmt.Println("Usage: " + os.Args[0] + " [create|sync]")
	fmt.Println("Make sure credentials file setup: http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs")
	os.Exit(1)
}

func main() {
	domainPtr := flag.String("domain", "", "Domain Name")
	flag.Parse()

	cmd := flag.Arg(0)

	if cmd == "create" {
		Create(*domainPtr)
	} else if cmd == "sync" {
		Sync()
	} else {
		Error()
	}
}

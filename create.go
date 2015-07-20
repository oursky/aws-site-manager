package main

import (
	"fmt"
	"io/ioutil"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/s3"
)

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
	DisplayAwsErr(err)
	fmt.Println(resp.ServerCertificateMetadata.ServerCertificateID)
	return *resp.ServerCertificateMetadata.ServerCertificateID
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

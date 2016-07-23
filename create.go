package main

import (
	"fmt"
	"io/ioutil"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
	fmt.Println(resp.ServerCertificateMetadata.ServerCertificateId)
	return *resp.ServerCertificateMetadata.ServerCertificateId
}

func Create(sess *session.Session, domain string, www bool, certID string) {
	svc := s3.New(sess)

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
		IAMCertificateId: aws.String(certID),
		SSLSupportMethod: aws.String("sni-only"),
	}

	distributionInput := &cloudfront.CreateDistributionInput{
		DistributionConfig: &cloudfront.DistributionConfig{
			Aliases: &cloudfront.Aliases{
				Quantity: aws.Int64(int64(quantityOfAliases)),
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
					QueryString: aws.Bool(false),
				},
				MinTTL:         aws.Int64(60),
				TargetOriginId: aws.String("S3-" + domain + "-SITE"),
				TrustedSigners: &cloudfront.TrustedSigners{
					Enabled:  aws.Bool(false),
					Quantity: aws.Int64(0),
				},
				ViewerProtocolPolicy: aws.String("allow-all"),
			},
			Enabled: aws.Bool(true),
			Origins: &cloudfront.Origins{
				Quantity: aws.Int64(1),
				Items: []*cloudfront.Origin{
					&cloudfront.Origin{
						DomainName: aws.String(domain + ".s3-website-" + *sess.Config.Region + ".amazonaws.com"),
						Id:         aws.String("S3-" + domain + "-SITE"),
						CustomOriginConfig: &cloudfront.CustomOriginConfig{
							HTTPPort:             aws.Int64(80),
							HTTPSPort:            aws.Int64(443),
							OriginProtocolPolicy: aws.String("http-only"),
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

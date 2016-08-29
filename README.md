# aws-site-manager
A very simple CLI to create S3 / Cloudfront Static Website. You can also reference [our blog](https://code.oursky.com/aws-site-manager-open-source-tool-hosting-static-web-on-s3-cloudfront/) for more details. 

<br>


## Install it

Binary downloads for Linux / Mac / Windows coming soon

If you have Go 1.6 or above installed, run the following command:

~~~~
go get -u github.com/oursky/aws-site-manager
go install github.com/oursky/aws-site-manager
~~~~

<br>

## How to use it?

**1. Set up AWS Credentials and config** 

If you haven't set up AWS credentials on your environment before, you shall set it up by putting the following lines in ~/.aws/credentials.
~~~~
[default]
AWS_ACCESS_KEY_ID=[MY_KEY_ID]
AWS_SECRET_ACCESS_KEY=[MY_SECRET_KEY]

~~~~
And in ~/.aws/config

~~~~
[default]
region=us-west-2
~~~~

You should also set the environment variable of AWS_SDK_LOAD_CONFIG (or put the following line in ~/.bashrc if you're on Linux / Mac assume you're using bash):

~~~~
export AWS_SDK_LOAD_CONFIG=1
~~~~
You can read more about AWS CLI set up on its [official documentation] (http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html).

<br>

**2. Use it!** 

Assume you're going to set up a website example.com and www.example.com, you can:
~~~~
cd ~/my_html_css_js
aws-site-manager create --domain example.com --www
aws-site-manager sync --domain example.com
~~~~
The command above will set up a S3 bucket example.com and www.example.com, sync all files in local folder, and redirect www.example.com to example.com

If you want to use https, and got the cert in PEM format ready (if your SSL registry sent you .key / .crt, read this)

~~~~
cd ~/my_html_css_js
aws-site-manager create --domain example.com --www --ssl --certBody body.pem --certChain chain.pem --privateKey key.pem
aws-site-manager sync --domain example.com
~~~~
After setting up the code above, you would need to set up Route53 or your DNS Manager to the CloudFront Distribution endpoint.

<br>

## What's next?

aws-site-manager is very preliminary, feel free to create issues if you run into any problems; Feel free to send us pull requests. And here are a list of things I'm working on:

* Support using gzip on cloudFront instead of S3
* Remember config so next time you can just run aws-site-manager sync on the local folder
* Better control on HTTP header, Custom page for Error code
* Support using Let's Encrypt free SSL cert or ACM cert
* Automatically configure Route-53 too

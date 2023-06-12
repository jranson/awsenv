# awsenv

This is a simple command line application that will parse AWS credentials from the clipboard and write them to a profile in `~/.aws/credentials`.

This utility is meant to be used with AWS IAM Identity Center (successor to AWS Single Sign-On).

## Usage

Navigate to your IAM Identity Center Login page (e.g., `https://d-XXXXXXXXXXXX.awsapps.com/`) and login. Identify the account and role you'd like to access and click 'Command line or programmatic access'. From the panel that appears, click on the quoted text in Option 1 or Option 2 to copy it to the clipboard.

Now run:
```bash
$ awsenv
```
This will import the clipboard data into the credentials file.

By default, data copied from Option 1 is stored in the `[default]` profile, while data from Option 2 is stored in the `[AcctNum_RoleName]` profile as provided in the credential data.

You can specify the target profile name as an extra command line parameter:

```bash
$ awsenv DevAccount
```

## Building and Installing

Ensure that the golang compiler is installed on your machine:

```bash
$ go version
```

Clone the project locally, and change working directory to the project root. Then run:

```bash
$ make build
```

Copy the binary from `./bin/awsenv` to the desired location in `$PATH`.

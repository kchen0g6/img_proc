# img_proc

Thanks for trying out this application. It runs locally on your machine so your
data is yours. You will provision resources on AWS via the Terraform file
provided. From there you will be able to perform image operations. Your results
will be stored in JSON and are downloadable.

So far you can perform two operation. One to compress images and one to read text.
Results are stored in `results/<job-id>.json` and also downloadable.

## Setup

### 1. Fork the repository and install Ruby dependencies

```bash
cd img_proc/app
bundle install
bin/rails db:create
```

### 2. Build the Lambda worker

Terraform packages this binary at plan time, so it has to exist before you
apply. 

```bash
cd ../go_proc
go mod download
GOOS=linux GOARCH=arm64 go build -o dist/bootstrap ./cmd/worker
```

### 3. Provision AWS

```bash
At this point ensure you have a sandbox account with environment variables or add key credentials to .tf

cd ../terraform
terraform init
terraform apply
terraform destroy
```

## Running

Start go and rails server on two terminals. 
Set concurrency flag (`-concurrency`) for the number of simultaneous Lambda invocations.

```bash
cd go_proc
AWS_REGION=us-east-1 go run ./cmd/ingest -concurrency=256
```

```bash
cd app
bin/rails s
```
Open <http://localhost:3000>, and perform desired action.

## Dataset

An example zip file is provided in dataset folder.

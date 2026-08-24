terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }

    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.4"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

locals {
  function_name = "img-proc-worker"
}

data "archive_file" "worker" {
  type        = "zip"
  source_file = "${path.module}/../go_proc/dist/bootstrap"
  output_path = "${path.module}/../go_proc/dist/worker.zip"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "worker" {
  name               = "${local.function_name}-role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "logs" {
  role       = aws_iam_role.worker.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

data "aws_iam_policy_document" "textract" {
  statement {
    actions   = ["textract:DetectDocumentText"]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "textract" {
  name   = "${local.function_name}-textract"
  role   = aws_iam_role.worker.id
  policy = data.aws_iam_policy_document.textract.json
}

resource "aws_cloudwatch_log_group" "worker" {
  name              = "/aws/lambda/${local.function_name}"
  retention_in_days = 3
}

resource "aws_lambda_function" "worker" {
  function_name = local.function_name
  role          = aws_iam_role.worker.arn

  filename         = data.archive_file.worker.output_path
  source_code_hash = data.archive_file.worker.output_base64sha256

  runtime       = "provided.al2023"
  handler       = "bootstrap"
  architectures = ["arm64"]

  memory_size = 1769 
  timeout     = 60

  depends_on = [
    aws_iam_role_policy_attachment.logs,
    aws_iam_role_policy.textract,
    aws_cloudwatch_log_group.worker,
  ]
}

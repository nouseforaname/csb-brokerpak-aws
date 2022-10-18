variable "aws_access_key_id" { type = string }
variable "aws_secret_access_key" { type = string }
variable "region" { type = string }
variable "instance_name" { type = string }
variable "aws_vpc_id" { type = string }
variable "cluster_instances" { type = number }
variable "serverless_min_capacity" { type = number }
variable "serverless_max_capacity" { type = number }
variable "engine_version" { type = string }
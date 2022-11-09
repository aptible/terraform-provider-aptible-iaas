data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

data "aws_vpc" "test_vpc" {
  depends_on = [aptible_aws_vpc.network]

  tags = {
    Name = var.vpc_name
  }
}

data "aws_subnets" "private_subnets" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.test_vpc.id]
  }

  tags = {
    Network = "Private"
  }
}

data "aws_subnets" "public_subnets" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.test_vpc.id]
  }

  tags = {
    Network = "Public"
  }
}

locals {
  private_subnet_id_to_use_for_instance = data.aws_subnets.private_subnets.ids[0]
  public_subnet_id_to_use_for_instance  = data.aws_subnets.public_subnets.ids[0]
}

resource "aws_security_group" "allow_access_to_igw" {
  name   = "allow_access_to_igw"
  vpc_id = data.aws_vpc.test_vpc.id

  ingress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_instance" "test_instance_private" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name = "test_inst_private_reach"
  }

  subnet_id = local.private_subnet_id_to_use_for_instance
}

resource "aws_network_interface_sg_attachment" "sg_attachment_private" {
  security_group_id    = aws_security_group.allow_access_to_igw.id
  network_interface_id = aws_instance.test_instance_private.primary_network_interface_id
}

resource "aws_instance" "test_instance_public" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"

  tags = {
    Name = "test_inst_public_reach"
  }

  subnet_id = local.public_subnet_id_to_use_for_instance
}

resource "aws_network_interface_sg_attachment" "sg_attachment_public" {
  security_group_id    = aws_security_group.allow_access_to_igw.id
  network_interface_id = aws_instance.test_instance_public.primary_network_interface_id
}

resource "aws_ec2_network_insights_path" "reachability_test" {
  source      = aws_instance.test_instance_private.primary_network_interface_id
  destination = aws_instance.test_instance_public.primary_network_interface_id
  protocol    = "tcp"
}

resource "aws_ec2_network_insights_analysis" "analysis" {
  network_insights_path_id = aws_ec2_network_insights_path.reachability_test.id
  wait_for_completion      = true
}

output "test_instance_private_arn" {
  value = aws_instance.test_instance_private.arn
}

output "test_instance_public_arn" {
  value = aws_instance.test_instance_public.arn
}

output "analysis_id" {
  value = aws_ec2_network_insights_analysis.analysis.id
}
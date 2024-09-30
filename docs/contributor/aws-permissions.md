# Google Cloud Platform Permissions

## Default Principal

```json
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"ec2:AssociateVpcCidrBlock",
				"ec2:AuthorizeSecurityGroupIngress",
				"ec2:CreateNetworkInterface",
				"ec2:CreateSecurityGroup",
				"ec2:CreateSubnet",
				"ec2:CreateTags",
				"ec2:DeleteNetworkInterface",
				"ec2:DeleteSecurityGroup",
				"ec2:DeleteSubnet",
				"ec2:DeleteTags",
				"ec2:DescribeNetworkInterfaceAttribute",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DescribeSecurityGroups",
				"ec2:DescribeSubnets",
				"ec2:DescribeVpcs",
				"ec2:DisassociateVpcCidrBlock",
				"ec2:RevokeSecurityGroupIngress",
				"elasticache:AddTagsToResource",
				"elasticache:CreateCacheCluster",
				"elasticache:CreateCacheParameterGroup",
				"elasticache:CreateCacheSubnetGroup",
				"elasticache:CreateReplicationGroup",
				"elasticache:CreateUserGroup",
				"elasticache:DeleteCacheCluster",
				"elasticache:DeleteCacheParameterGroup",
				"elasticache:DeleteCacheSubnetGroup",
				"elasticache:DeleteReplicationGroup",
				"elasticache:DeleteUserGroup",
				"elasticache:DescribeCacheClusters",
				"elasticache:DescribeCacheParameterGroups",
				"elasticache:DescribeCacheParameters",
				"elasticache:DescribeCacheSubnetGroups",
				"elasticache:DescribeEngineDefaultParameters",
				"elasticache:DescribeReplicationGroups",
				"elasticache:DescribeUserGroups",
				"elasticache:ListTagsForResource",
				"elasticache:ModifyCacheCluster",
				"elasticache:ModifyCacheParameterGroup",
				"elasticache:ModifyReplicationGroup",
				"elasticache:RemoveTagsFromResource",
				"elasticfilesystem:CreateFileSystem",
				"elasticfilesystem:CreateMountTarget",
				"elasticfilesystem:CreateTags",
				"elasticfilesystem:DeleteFileSystem",
				"elasticfilesystem:DeleteMountTarget",
				"elasticfilesystem:DeleteTags",
				"elasticfilesystem:DescribeFileSystems",
				"elasticfilesystem:DescribeMountTargetSecurityGroups",
				"elasticfilesystem:DescribeMountTargets",
				"elasticfilesystem:TagResource",
				"elasticfilesystem:UntagResource",
				"secretsmanager:CreateSecret",
				"secretsmanager:DeleteSecret",
				"secretsmanager:DescribeSecret",
				"secretsmanager:GetSecretValue",
				"secretsmanager:TagResource",
				"sts:GetCallerIdentity"
			],
			"Resource": "*"
		},
		{
			"Effect": "Allow",
			"Action": [
				"iam:CreateServiceLinkedRole",
				"iam:PutRolePolicy",
				"iam:DeleteServiceLinkedRole",
				"iam:GetServiceLinkedRoleDeletionStatus"
			],
			"Resource": "arn:aws:iam::*:role/aws-service-role/elasticache.amazonaws.com/AWSServiceRoleForElastiCache*",
			"Condition": {
				"StringLike": {
					"iam:AWSServiceName": "elasticache.amazonaws.com"
				}
			}
		}
	]
}
```

## Peering Principal
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:CreateRoute",
                "ec2:CreateTags",
                "ec2:CreateVpcPeeringConnection",
                "ec2:DeleteRoute",
                "ec2:DeleteVpcPeeringConnection",
                "ec2:DescribeRouteTables",
                "ec2:DescribeVpcPeeringConnections",
                "ec2:DescribeVpcs",
                "sts:GetCallerIdentity"
            ],
            "Resource": "*"
        }
    ]
}

```


# Create Virtual Private Cloud Peering in Amazon Web Services

This tutorial explains how to create a Virtual Private Cloud (VPC) peering connection between a remote VPC network and Kyma in Amazon Web Services (AWS). Learn how to create a new VPC network, and a virtual machine (VM), and assign required permissions to the provided Kyma account and role in your AWS account.

## Prerequisites <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* The AWS CLI configured. For instructions, see the [AWS documentation](https://docs.aws.amazon.com/cli/v1/userguide/cli-chap-configure.html).

## Steps <!-- {docsify-ignore} -->

1. Set the default AWS CLI profile.

    ```shell
    export AWS_PROFILE={PROFILE_NAME}
    export AWS_REGION={REGION}
    ```

2. Create a trust policy document.

    ```shell
    export PRINCIPAL_PROFILE_AWS_ACCOUNT_ID=194230256199
    export USER_NAME=cloud-manager-peering-dev
    cat > trust_policy.json <<- EOF
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "AWS": "arn:aws:iam::$PRINCIPAL_PROFILE_AWS_ACCOUNT_ID:user/$USER_NAME"
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }
    EOF
    ```

3. Create **CloudManagerPeeringRole** and attach a trust policy document.

    ```shell
    export AWS_ROLE_NAME=CloudManagerPeeringRole
    aws iam create-role --role-name $AWS_ROLE_NAME --assume-role-policy-document file://./trust_policy.json 
    ```

4. Create a policy document that is used to create the policy.

    ```shell
    cat > accept_policy.json <<- EOF
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "ec2:AcceptVpcPeeringConnection",
                    "ec2:DescribeVpcs",
                    "ec2:DescribeVpcPeeringConnections",
                    "ec2:DescribeRouteTables",
                    "ec2:CreateRoute",
                    "ec2:CreateTags"
                ],
                "Resource": "*"
            }
        ]
    }
    EOF
    ```

5. Create a new managed policy for your Amazon Web Services account.

    ```shell
    aws iam create-policy --policy-name CloudManagerPeeringAccess --policy-document file://./accept_policy.json
    ```

6. Attach the specified managed policy to the specified IAM role.

    ```shell
    aws iam attach-role-policy --role-name $AWS_ROLE_NAME --policy-arn arn:aws:iam::$REMOTE_ACCOUNT_ID:policy/CloudManagerPeeringAccess
    ```

7. Create a VPC and tag it with a Kyma shoot name.

    ```shell
    export CIDR_BLOCK=10.3.0.0/16
    export SHOOT_NAME=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.shootName}')
    export NODE_NETWORK=$(kubectl get cm -n kube-system shoot-info -o jsonpath='{.data.nodeNetwork}')
    export VPC_NAME=my-vpc
    export VPC_ID=$(aws ec2 create-vpc --cidr-block $CIDR_BLOCK --tag-specifications "ResourceType=vpc,Tags=[{Key=$SHOOT_NAME,Value=''},{Key=Name,Value=$VPC_NAME}]" --query Vpc.VpcId --output text)
    ```

8. Find an availability zone that supports the instance type compatible with the specified image.

    ```shell
    export IMAGE_ID=$(aws ec2 describe-images --owners amazon --filters "Name=name,Values=ubuntu/images/hvm-ssd-gp3/ubuntu-noble*" --query 'sort_by(Images, &CreationDate)[-1].ImageId' --output text)
    export INSTANCE_TYPE=$(aws ec2 describe-instance-types --filter "Name=instance-type,Values=*.micro" "Name=processor-info.supported-architecture,Values=arm64" --query 'InstanceTypes[0].InstanceType' | tr -d '"')
    export AVAILABILITY_ZONE=$(aws ec2 describe-instance-type-offerings --location-type availability-zone --filters "Name=instance-type,Values=$INSTANCE_TYPE" --query 'InstanceTypeOfferings[0].Location' --output text)
    ```

9. Create a subnet.

    ```shell
    export SUBNET_ID=$(aws ec2 create-subnet --vpc-id $VPC_ID --availability-zone $AVAILABILITY_ZONE --cidr-block $CIDR_BLOCK --query Subnet.SubnetId --output text) 
    ```

10. Run an instance.

     ```shell
     export INSTANCE_ID=$(aws ec2 run-instances --image-id $IMAGE_ID --instance-type $INSTANCE_TYPE --subnet-id $SUBNET_ID --query "Instances[0].InstanceId" --output text)
     export IP_ADDRESS=$(aws ec2 describe-instances --instance-ids $INSTANCE_ID --query "Reservations[0].Instances[0].PrivateIpAddress" --output text)
     ```

11. Allow ICMP traffic from Kyma Pods.

    ```shell
     export SG_ID=$(aws ec2 describe-security-groups --filters Name=vpc-id,Values=$VPC_ID --query "SecurityGroups[0].GroupId" --output text) 
     aws ec2 authorize-security-group-ingress --group-id $SG_ID --ip-permissions IpProtocol=icmp,FromPort=-1,ToPort=-1,IpRanges="[{CidrIp=$NODE_NETWORK}]"
    ```

12. Create an AwsVpcPeering resource.

    ```shell
    export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    kubectl apply -f - <<EOF
    apiVersion: cloud-resources.kyma-project.io/v1beta1
    kind: AwsVpcPeering
    metadata:
      name: peering-to-my-vpc
    spec:
      remoteAccountId: "$ACCOUNT_ID"
      remoteRegion: "$AWS_REGION"
      remoteVpcId: "$VPC_ID"
    EOF
    ```

13. Wait for the AwsVpcPeering to be in the `Ready` state.

    ```shell
    kubectl wait --for=condition=Ready awsvpcpeering/peering-to-my-vpc --timeout=300s
    ```

    Once the newly created AwsVpcPeering is provisioned, you should see the following message:

    ```txt
    awsvpcpeering.cloud-resources.kyma-project.io/peering-to-my-vpc condition met
    ```

14. Create a namespace and export its value as an environment variable. Run:

    ```shell
    export NAMESPACE={NAMESPACE_NAME}
    kubectl create ns $NAMESPACE
    ```

15. Create a workload that pings the VM in the remote network.

    ```shell
    kubectl apply -n $NAMESPACE -f - <<EOF
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: awsvpcpeering-demo
    spec:
      selector:
        matchLabels:
          app: awsvpcpeering-demo
      template:
        metadata:
          labels:
            app: awsvpcpeering-demo
        spec:
          containers:
          - name: my-container
            resources:
              limits:
                memory: 512Mi
                cpu: "1"
              requests:
                memory: 256Mi
                cpu: "0.2"
            image: ubuntu
            command:
              - "/bin/bash"
              - "-c"
              - "--"
            args:
             - "apt update; apt install iputils-ping -y; ping -c 20 $IP_ADDRESS"
    EOF
    ```

    This workload should print a sequence of 20 echo replies to stdout.

16. To print the logs of one of the workloads, run:

    ```shell
    kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=awsvpcpeering-demo -o=jsonpath='{.items[0].metadata.name}'`
    ```

    The command should print an output similar to the following:

    ```txt
    ...
    PING 172.0.0.4 (172.0.0.4) 56(84) bytes of data.
    64 bytes from 172.0.0.4: icmp_seq=1 ttl=63 time=8.10 ms
    64 bytes from 172.0.0.4: icmp_seq=2 ttl=63 time=2.01 ms
    64 bytes from 172.0.0.4: icmp_seq=3 ttl=63 time=7.02 ms
    64 bytes from 172.0.0.4: icmp_seq=4 ttl=63 time=1.87 ms
    64 bytes from 172.0.0.4: icmp_seq=5 ttl=63 time=1.89 ms
    64 bytes from 172.0.0.4: icmp_seq=6 ttl=63 time=4.75 ms
    64 bytes from 172.0.0.4: icmp_seq=7 ttl=63 time=2.01 ms
    64 bytes from 172.0.0.4: icmp_seq=8 ttl=63 time=4.26 ms
    64 bytes from 172.0.0.4: icmp_seq=9 ttl=63 time=1.89 ms
    64 bytes from 172.0.0.4: icmp_seq=10 ttl=63 time=2.08 ms
    64 bytes from 172.0.0.4: icmp_seq=11 ttl=63 time=2.01 ms
    64 bytes from 172.0.0.4: icmp_seq=12 ttl=63 time=2.24 ms
    64 bytes from 172.0.0.4: icmp_seq=13 ttl=63 time=1.80 ms
    64 bytes from 172.0.0.4: icmp_seq=14 ttl=63 time=4.32 ms
    64 bytes from 172.0.0.4: icmp_seq=15 ttl=63 time=2.03 ms
    64 bytes from 172.0.0.4: icmp_seq=16 ttl=63 time=2.03 ms
    64 bytes from 172.0.0.4: icmp_seq=17 ttl=63 time=5.19 ms
    64 bytes from 172.0.0.4: icmp_seq=18 ttl=63 time=1.86 ms
    64 bytes from 172.0.0.4: icmp_seq=19 ttl=63 time=1.92 ms
    64 bytes from 172.0.0.4: icmp_seq=20 ttl=63 time=1.92 ms

    === 172.0.0.4 ping statistics ===
    20 packets transmitted, 20 received, 0% packet loss, time 19024ms
    rtt min/avg/max/mdev = 1.800/3.060/8.096/1.847 ms
    ...
    ```

## Next Steps

To clean up the Kubernetes resources and your AWS account resources, follow these steps:

1. Remove the created workloads:

   ```shell
   kubectl delete -n $NAMESPACE deployment awsvpcpeering-demo
   ```

2. Remove the created AwsVpcPeering:

   ```shell
   kubectl delete -n $NAMESPACE awsvpcpeering peering-to-my-vpc
   ```

3. Remove the created namespace:

   ```shell
   kubectl delete namespace $NAMESPACE
   ```

4. Go to your AWS account and terminate the instance.

   ```shell
   aws ec2 terminate-instances --instance-ids $INSTANCE_ID
   aws ec2 wait instance-terminated --instance-ids $INSTANCE_ID
   ```

5. Delete the subnet.

   ```shell
   aws ec2 delete-subnet --subnet-id $SUBNET_ID
   ```

6. Delete the VPC.

   ```shell
   aws ec2 delete-vpc --vpc-id  $VPC_ID
   ```
